package controller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/http/helper"
	"github.com/artufi/trader/http/middleware"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/artufi/trader/xtb/processor"
	"github.com/artufi/trader/xtb/response"
	"log/slog"
	"net/http"
	"strconv"
)

type Purchase struct {
	Cfg       config.AppConfig
	WSManager *websocket.WSManager
	Logger    *slog.Logger

	PredictionService model.PredictionService
	OrderService      model.OrderService
	UserService       model.UserService
}

func (p *Purchase) PurchaseHandler(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())
	userID := p.Cfg.XTB.Demo.UserID
	ctx := logging.AppendAttrsCtx(r.Context(), logging.TraceIDAttr(traceID), logging.UserIDAttr(userID))

	p.Logger.InfoContext(ctx, "Start processing purchase request")

	rh := helper.Response{Logger: p.Logger, Writer: w}

	var prediction model.PredictionDetails
	if err := json.NewDecoder(r.Body).Decode(&prediction); err != nil {
		rh.WriteAndLogError(ctx, http.StatusBadRequest, "Failed to decode prediction", err)
		return
	}

	p.processPrediction(ctx, prediction, userID, traceID, rh)
	w.WriteHeader(http.StatusNoContent)
	p.Logger.InfoContext(ctx, "Purchases request fully processed")
}

func (p *Purchase) PurchasesHandler(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())
	userID := p.Cfg.XTB.Demo.UserID
	ctx := logging.AppendAttrsCtx(r.Context(), logging.TraceIDAttr(traceID), logging.UserIDAttr(userID))

	rh := helper.Response{Logger: p.Logger, Writer: w}

	p.Logger.InfoContext(ctx, "Start processing purchases request")

	var predictions []model.PredictionDetails
	if err := json.NewDecoder(r.Body).Decode(&predictions); err != nil {
		rh.WriteAndLogError(ctx, http.StatusBadRequest, "Failed to decode predictions", err)
		return
	}

	for _, prediction := range predictions {
		p.processPrediction(ctx, prediction, userID, traceID, rh)
	}

	w.WriteHeader(http.StatusNoContent)
	p.Logger.InfoContext(ctx, "Purchases request fully processed")
}

func (p *Purchase) processPrediction(ctx context.Context, predictionDetails model.PredictionDetails, userID string,
	traceID string, rh helper.Response) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		rh.WriteAndLogError(ctx, http.StatusBadRequest, "Invalid userID", err)
		return
	}

	user, err := p.UserService.SelectUserByUsername(userIDInt)
	if err != nil {
		rh.WriteAndLogError(ctx, http.StatusNotFound, "User lookup failed", err)
		return
	}

	wsClient, err := p.WSManager.GetUserRandomClient(userID)
	if err != nil {
		rh.WriteAndLogError(ctx, http.StatusInternalServerError, "No websocket client", err)
		return
	}

	ctx = logging.AppendAttrsCtx(ctx,
		logging.ConnNo(wsClient.ConnID),
		logging.StreamID(wsClient.StreamSessionID),
		logging.SymbolAttr(predictionDetails.Symbol),
	)

	apiH := APIHandler{
		Proc:    processor.NewProc(wsClient),
		TraceID: traceID,
	}
	dbH := DBHandler{
		TraceID:           traceID,
		Logger:            p.Logger,
		PredictionService: p.PredictionService,
		OrderService:      p.OrderService,
	}

	symbolResp, err := apiH.GetSymbolExtended(ctx, predictionDetails.Symbol)
	if err != nil {
		rh.WriteAndLogError(ctx, http.StatusInternalServerError, "GetSymbolExtended failed", err)
		return
	}
	p.Logger.InfoContext(ctx, "Successfully retrieved symbol data", logging.RespAttr(symbolResp))

	tradeTransInfo, err := predictionDetails.PrepareTradeTransInfo(symbolResp, p.Cfg.BuySellParams.Volume)
	if err != nil {
		if errors.Is(err, model.ErrNoAction) {
			p.Logger.InfoContext(ctx, "ModelType: NoAction - stop processing")
		} else {
			p.Logger.WarnContext(ctx, "Prepare TradeTransInfo failed", logging.ErrorAttr(err))
		}
		return
	}
	p.Logger.InfoContext(ctx, "Prepared TradeTransInfo", logging.RespAttr(tradeTransInfo))

	tradeTransResp, err := apiH.TradeTransaction(ctx, predictionDetails.Symbol, tradeTransInfo)
	if err != nil {
		rh.WriteAndLogError(ctx, http.StatusInternalServerError, "TradeTransaction failed", err)
		return
	}
	p.Logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeTransResp))

	tradeTransStatusResp, err := apiH.TradeTransactionStatus(ctx, predictionDetails.Symbol, tradeTransResp.ReturnData.Order)
	if err != nil {
		rh.WriteAndLogError(ctx, http.StatusInternalServerError, "TradeTransactionStatus failed", err)
		return
	}
	p.Logger.InfoContext(ctx, "Successfully retrieved TradeTransactionStatus", logging.RespAttr(tradeTransStatusResp))

	prediction := model.Prediction{
		UserID:            user.ID,
		PredictionDetails: predictionDetails,
	}

	predID, err := dbH.InsertPrediction(ctx, prediction)
	if err != nil {
		p.Logger.ErrorContext(ctx, "Failed to insert prediction", logging.ErrorAttr(err))
		return
	}
	p.Logger.InfoContext(ctx, "Inserted prediction", logging.IDAttr(predID))

	reqStatus, errReq := tradeTransStatusResp.CheckRequestStatus()
	var order model.Order
	if reqStatus != response.REJECTED {
		// getTrades to acquire position required to forcibly close purchase on demand in history service.
		getTradesResp, err := apiH.GetTrades(ctx, predictionDetails.Symbol, true)
		if err != nil {
			p.Logger.WarnContext(ctx, "Failed to process GetTrades, position number will not be added", logging.ErrorAttr(err))
		} else {
			p.Logger.InfoContext(ctx, "Successfully processed GetTrades", logging.RespAttr(getTradesResp))
		}

		var position int
		for _, tradeRecord := range getTradesResp.ReturnData {
			// Each transaction in the XTB system is assigned a unique order number. When a TradeTransaction
			// is executed by the client, the system returns an order number for that transaction.
			// Subsequently, request made by the client is processed by the system, which marks the start
			// of a new transaction in their system. As a result, the previous order number is transformed
			// into 'order2', indicating that the previous system transaction has simply ended,
			// and a new transaction with a different order number is being processed. When the previous
			// order number matches the 'order2' number retrieved from XTB, it indicates that transactions
			// in their system are connected, and it is possible to obtain a position number that represents
			// the entire sequence of transactions.
			//
			// This process resembles a chain of transactions: a TradeTransaction from the client assigns
			// an order number to that transaction, then the system processes the TradeTransaction,
			// assigns it a new order number, and saves the previous order number of transaction in the
			// 'order2' field. Finally, it returns the position number that represents the complete
			// chain of transactions in the system.
			if tradeTransResp.ReturnData.Order == tradeRecord.Order2 {
				position = tradeRecord.Position
				p.Logger.InfoContext(ctx, "Acquired order position", logging.PositionAttr(position))
				break
			}
		}

		order = model.Order{
			Number:         tradeTransResp.ReturnData.Order,
			UserID:         user.ID,
			Position:       &position,
			PredictionID:   predID,
			RequestStatus:  response.RequestStatusName[reqStatus],
			Message:        tradeTransStatusResp.ReturnData.Message,
			TradeTransInfo: tradeTransInfo,
		}
	} else {
		rsErr := &response.RequestStatusError{}
		if errors.As(errReq, &rsErr) {
			order = model.Order{
				Number:         tradeTransResp.ReturnData.Order,
				UserID:         user.ID,
				PredictionID:   predID,
				RequestStatus:  rsErr.RequestStatus,
				Message:        &rsErr.Message,
				TradeTransInfo: tradeTransInfo,
				OrderClosedDetails: model.OrderClosedDetails{
					Closed: true,
				},
			}
			p.Logger.InfoContext(ctx, "Order rejected, marked as closed", logging.OrderAttr(order.Number))
		}
	}

	ordID, err := dbH.InsertOrder(ctx, order)
	if err != nil {
		p.Logger.ErrorContext(ctx, "Insert order failed", logging.ErrorAttr(err))
	}
	p.Logger.InfoContext(ctx, "Inserted order", logging.IDAttr(ordID))
}
