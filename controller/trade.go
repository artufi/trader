package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"trader/config"
	"trader/controller/middleware"
	"trader/infrastructure/websocket"
	"trader/logging"
	"trader/model"
	"trader/xtb/command"
	jsonxtb "trader/xtb/json"
	"trader/xtb/response"
)

func TradeHandler(cfg config.AppConfig, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "Start processing trade request")

		connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails)
		if !ok {
			http.Error(w, "Connection details are not available", http.StatusBadRequest)
			return
		}

		wsClient, err := wsManager.DialForNewClient(ctx, cfg.XTB.Demo.WebSocketURL, nil)
		if err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO
		// size 1?
		respCh := make(chan []byte, 1)
		go wsClient.ReadMessages(ctx, respCh)

		loginJSON, err := jsonxtb.Login(command.LoginArgs{
			UserID:   connDetails.UserID,
			Password: cfg.Testing.Password,
		})
		if err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Info("Processing message", logging.MsgAttr(string(loginJSON)))
		err = wsClient.WriteText(ctx, loginJSON)
		if err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		loginResponse, ok := <-respCh
		if !ok {
			wsManager.RemoveClient(ctx, wsClient)
			logger.ErrorContext(ctx, "Channel closed something happened with reader while processing loginResponse")
			return
		}
		wsLoginResponse := response.GeneralResponse{}
		if err := json.Unmarshal(loginResponse, &wsLoginResponse); err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			logger.ErrorContext(ctx, "Failed to read a response", logging.ErrorAttr(err))
			return
		}
		logger.Info("Received response", "resp", wsLoginResponse)

		// get trade instructions
		tradeInstrs := make([]model.TradeInstruction, 0)
		err = json.NewDecoder(r.Body).Decode(&tradeInstrs)
		if err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		errors := make([]string, 0)
		for _, tradeInstr := range tradeInstrs {
			logger.InfoContext(ctx, "Processing trade instruction", logging.TradeInstrAttr(tradeInstr))

			// get symbol details
			symbol := strings.ToUpper(tradeInstr.Symbol)
			symbolJSON, err := jsonxtb.GetSymbol(command.GetSymbolArgs{
				Symbol: symbol,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process getSymbol", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				errors = append(errors, err.Error())
				continue
			}
			logger.Info("Processing message", logging.MsgAttr(string(symbolJSON)))
			err = wsClient.WriteText(ctx, symbolJSON)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to write text", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				errors = append(errors, err.Error())
				continue
			}

			symbolResponse := <-respCh
			wsSymbolResponse := response.GetSymbol{}
			if err := json.Unmarshal(symbolResponse, &wsSymbolResponse); err != nil {
				wsManager.RemoveClient(ctx, wsClient)
				logger.ErrorContext(ctx, "Failed to read a response", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				errors = append(errors, err.Error())
				return
			}
			logger.Info("Received response", "resp", wsSymbolResponse)

			tradeTransInfo := command.TradeTransInfo{
				CustomComment: "test",
				Expiration:    time.Now().Add(time.Minute * 10).UnixMilli(),
				Price:         wsSymbolResponse.ReturnData.Ask,
				Symbol:        symbol,
				Type:          command.BUY,
				Volume:        0.1,
				// precision!
				Tp: tradeInstr.TakeProfit,
			}
			if tradeInstr.PredsProba >= 0.5 {
				tradeTransInfo.Cmd = command.OPEN
			} else {
				logger.InfoContext(ctx, "Too low value nothing to process",
					logging.SymbolAttr(symbol), "preds_proba", tradeInstr.PredsProba)
				continue
			}
			tradeTransactionJSON, err := jsonxtb.TradeTransaction(command.TradeTransactionArgs{
				TradeTransInfo: tradeTransInfo})

			if err != nil {
				logger.ErrorContext(ctx, "Failed to process tradeTransaction", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				errors = append(errors, err.Error())
				continue
			}
			logger.InfoContext(ctx, "Message to process", logging.MsgAttr(string(tradeTransactionJSON)))
			err = wsClient.WriteText(ctx, tradeTransactionJSON)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to write text", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				errors = append(errors, err.Error())
				continue
			}
			tradeTransResponse := <-respCh
			wsTradeTransResponse := response.GeneralResponse{}
			if err := json.Unmarshal(tradeTransResponse, &wsTradeTransResponse); err != nil {
				wsManager.RemoveClient(ctx, wsClient)
				logger.ErrorContext(ctx, "Failed to read a response", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				errors = append(errors, err.Error())
				return
			}
			logger.Info("Received response", logging.SymbolAttr(symbol), "resp", wsTradeTransResponse)
		}

		if len(errors) > 0 {
			errorsJoined := strings.Join(errors, ",")
			logger.ErrorContext(ctx, "Failed to write text", "errors", errorsJoined)
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, errorsJoined, http.StatusInternalServerError)
			return
		}
		logger.Info("Successfully processed")
	}
}
