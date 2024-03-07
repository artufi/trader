package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

func TradeTransaction(ctx context.Context, tradeTransInfo command.TradeTransInfo, wsClient *websocket.WSClient, customTag string) (response.TradeTransaction, error) {
	tradeTransactionJSON, err := jsonform.TradeTransaction(tradeTransInfo, customTag)
	if err != nil {
		return response.TradeTransaction{}, fmt.Errorf("XTB TradeTransaction processor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, customTag, tradeTransactionJSON)
	if err != nil {
		return response.TradeTransaction{}, fmt.Errorf("XTB TradeTransaction processor: %w", err)
	}

	wsTradeTransResp := response.TradeTransaction{}
	if err := json.Unmarshal(resp, &wsTradeTransResp); err != nil {
		return response.TradeTransaction{}, fmt.Errorf("XTB TradeTransaction processor: %w", err)
	}
	return wsTradeTransResp, err
}

func TradeTransactionStatus(ctx context.Context, orderNo int, wsClient *websocket.WSClient, customTag string) (response.TradeTransactionStatus, error) {
	tradeTransactionStatusJSON, err := jsonform.TradeTransactionStatus(orderNo, customTag)
	if err != nil {
		return response.TradeTransactionStatus{}, fmt.Errorf("XTB TradeTransactionStatus processor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, customTag, tradeTransactionStatusJSON)
	if err != nil {
		return response.TradeTransactionStatus{}, fmt.Errorf("XTB TradeTransactionStatus processor: %w", err)
	}

	wsTradeTransResp := response.TradeTransactionStatus{}
	if err := json.Unmarshal(resp, &wsTradeTransResp); err != nil {
		return response.TradeTransactionStatus{}, fmt.Errorf("XTB TradeTransactionStatus processor: %w", err)
	}
	return wsTradeTransResp, err
}
