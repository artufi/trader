package history

import (
	"context"
	"errors"
	"github.com/artufi/trader/http/controller"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/processor"
	"github.com/artufi/trader/xtb/response"
	"github.com/google/uuid"
	"time"
)

const closeServiceName = "CloseService"

func (hs *HService) StartOrderCloseLoop(ctx context.Context, userID string, interval time.Duration) {
	ctx = logging.AppendAttrsCtx(ctx, logging.ServiceName(closeServiceName), logging.UserIDAttr(userID))

	ticker := time.NewTicker(time.Second * interval)
	defer ticker.Stop()

	hs.Logger.InfoContext(ctx, "Starting order close loop", logging.IntervalAttr(interval))

	for {
		select {
		case <-ticker.C:
			hs.closeEligibleOrders(ctx, userID)
		case <-ctx.Done():
			hs.Logger.InfoContext(ctx, "Context done, stop closing eligible orders", logging.ErrorAttr(ctx.Err()))
			return
		}
	}
}

func (hs *HService) closeEligibleOrders(ctx context.Context, userID string) {
	wsClient, err := hs.WSManager.GetUserRandomClient(userID)
	if err != nil {
		hs.Logger.ErrorContext(ctx, "Failed to get client to process closing orders", logging.ErrorAttr(err))
		return
	}

	apiH := controller.APIHandler{
		Proc:    processor.NewProc(wsClient),
		TraceID: uuid.New().String(),
	}

	ordersWithPosition, err := hs.OrderService.SelectOpenOrdersWithPosition()
	if err != nil {
		hs.Logger.ErrorContext(ctx, "Failed to get open orders with position from database", logging.ErrorAttr(err))
	}

	if len(ordersWithPosition) > 0 {
		hs.Logger.InfoContext(ctx, "Will process order with position")
		hs.closeOrdersWithPosition(ctx, apiH, ordersWithPosition)
	} else {
		hs.Logger.InfoContext(ctx, "Will process order without position")
		hs.closeOrdersWithoutPosition(ctx, apiH)
	}
}

func (hs *HService) closeOrdersWithPosition(ctx context.Context, apiH controller.APIHandler, ordersWithPosition []model.Order) {
	for _, order := range ordersWithPosition {
		symbol := order.Symbol
		ctx := logging.AppendAttrsCtx(ctx, logging.SymbolAttr(symbol))
		symbolResponse, err := apiH.GetSymbolExtended(ctx, symbol)
		if err != nil {
			hs.Logger.ErrorContext(ctx, "Failed to process GetSymbol", logging.ErrorAttr(err))
			continue
		}
		hs.Logger.InfoContext(ctx, "Successfully processed GetSymbol", logging.RespAttr(symbolResponse))

		tradeTransInfo := command.TradeTransInfo{
			CustomComment: "CLOSE TRANSACTION",
			Expiration:    time.Now().Add(time.Minute).UnixMilli(),
			Order:         order.Position,
			Price:         symbolResponse.ReturnData.Ask,
			Symbol:        symbol,
			Type:          command.CLOSE,
			Volume:        order.Volume,
		}
		tradeTransResp, err := apiH.TradeTransaction(ctx, symbol, tradeTransInfo)
		if err != nil {
			hs.Logger.ErrorContext(ctx, "Failed to process TradeTransaction", logging.ErrorAttr(err))

			var statusError *response.StatusError
			if errors.As(err, &statusError) && statusError.ErrorCode == "SE199" {
				hs.Logger.ErrorContext(ctx, "Probably Order is already closed but status is not refreshed in a database",
					logging.OrderAttr(order.Number),
					logging.PositionAttr(*order.Position))

				id, err := hs.OrderService.MarkFailedAsClosed(order.Number)
				if err != nil {
					hs.Logger.ErrorContext(ctx, "Failed to update failed order", logging.ErrorAttr(err))
				} else {
					hs.Logger.InfoContext(ctx, "Updated order", logging.IDAttr(id))
				}
			}
			continue
		}
		hs.Logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeTransResp))

		tradeTransStatusResp, err := apiH.TradeTransactionStatus(ctx, symbol, tradeTransResp.ReturnData.Order)
		if err != nil {
			hs.Logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
			continue
		}
		hs.Logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeTransStatusResp))

		hs.Logger.InfoContext(ctx, "Close 'order' request fully processed", logging.OrderAttr(order.Number), logging.PositionAttr(*order.Position))
	}
}

func (hs *HService) closeOrdersWithoutPosition(ctx context.Context, apiH controller.APIHandler) {
	orders, err := hs.OrderService.SelectOpenOrders()
	if err != nil {
		hs.Logger.ErrorContext(ctx, "Failed to get open orders from database", logging.ErrorAttr(err))
		return
	}
	hs.Logger.InfoContext(ctx, "Successfully processed selecting orders", logging.RespAttr(orders))

	orderMap := make(map[int][]model.Order)
	for _, order := range orders {
		orderMap[order.Number] = append(orderMap[order.Number], order)
	}

	// getTrades to acquire position required to forcibly close purchase on demand in history service.
	getTradesResp, err := apiH.GetTrades(ctx, "", true)
	if err != nil {
		hs.Logger.WarnContext(ctx, "Failed to process GetTrades",
			logging.ErrorAttr(err))
	} else {
		hs.Logger.InfoContext(ctx, "Successfully processed GetTrades", logging.RespAttr(getTradesResp))
	}

	// TODO: would be good to update an order with a position number.
	for _, tradeRecord := range getTradesResp.ReturnData {
		if !tradeRecord.Closed {
			matchedOrders := orderMap[tradeRecord.Order2]
			if len(matchedOrders) == 0 {
				continue
			}
			for _, order := range matchedOrders {
				symbol := order.Symbol
				ctx := logging.AppendAttrsCtx(ctx, logging.SymbolAttr(symbol))
				symbolResponse, err := apiH.GetSymbolExtended(ctx, symbol)
				if err != nil {
					hs.Logger.ErrorContext(ctx, "Failed to process GetSymbol", logging.ErrorAttr(err))
					continue
				}
				hs.Logger.InfoContext(ctx, "Successfully processed GetSymbol", logging.RespAttr(symbolResponse))

				tradeTransInfo := command.TradeTransInfo{
					CustomComment: "CLOSE TRANSACTION",
					Expiration:    time.Now().Add(time.Minute).UnixMilli(),
					Order:         &tradeRecord.Position,
					Price:         symbolResponse.ReturnData.Ask,
					Symbol:        symbol,
					Type:          command.CLOSE,
					Volume:        order.Volume,
				}
				tradeTransResp, err := apiH.TradeTransaction(ctx, symbol, tradeTransInfo)
				if err != nil {
					hs.Logger.ErrorContext(ctx, "Failed to process TradeTransaction", logging.ErrorAttr(err))

					var statusError *response.StatusError
					if errors.As(err, &statusError) && statusError.ErrorCode == "SE199" {
						hs.Logger.ErrorContext(ctx, "Probably Order is already closed but status is not refreshed in a database",
							logging.OrderAttr(order.Number),
							logging.PositionAttr(*order.Position))

						id, err := hs.OrderService.MarkFailedAsClosed(order.Number)
						if err != nil {
							hs.Logger.ErrorContext(ctx, "Failed to update failed order", logging.ErrorAttr(err))
						} else {
							hs.Logger.InfoContext(ctx, "Updated order", logging.IDAttr(id))
						}
					}
					continue
				}
				hs.Logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeTransResp))

				tradeTransStatusResp, err := apiH.TradeTransactionStatus(ctx, symbol, tradeTransResp.ReturnData.Order)
				if err != nil {
					hs.Logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
					continue
				}
				hs.Logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeTransStatusResp))

				hs.Logger.InfoContext(ctx, "Close 'order' request fully processed", logging.OrderAttr(order.Number), logging.PositionAttr(*order.Position))
			}
		}
	}
}
