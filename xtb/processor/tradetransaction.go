package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"trader/infrastructure/websocket"
	"trader/xtb/command"
	"trader/xtb/jsonform"
	"trader/xtb/response"
)

func TradeTransaction(ctx context.Context, tradeTransInfo command.TradeTransInfo, wsClient *websocket.WSClient) (response.TradeTransaction, error) {
	tradeTransactionJSON, err := jsonform.TradeTransaction(command.TradeTransactionArgs{
		TradeTransInfo: tradeTransInfo})
	if err != nil {
		return response.TradeTransaction{}, fmt.Errorf("XTB TradeTransaction processor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, tradeTransactionJSON)
	if err != nil {
		return response.TradeTransaction{}, fmt.Errorf("XTB TradeTransaction processor: %w", err)
	}

	wsTradeTransResp := response.TradeTransaction{}
	if err := json.Unmarshal(resp, &wsTradeTransResp); err != nil {
		return response.TradeTransaction{}, fmt.Errorf("XTB TradeTransaction processor: %w", err)
	}
	return wsTradeTransResp, err
}

func TradeTransactionStatus(ctx context.Context, orderNo int, wsClient *websocket.WSClient) (response.TradeTransactionStatus, error) {
	tradeTransactionStatusJSON, err := jsonform.TradeTransactionStatus(orderNo)
	if err != nil {
		return response.TradeTransactionStatus{}, fmt.Errorf("XTB TradeTransactionStatus processor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, tradeTransactionStatusJSON)
	if err != nil {
		return response.TradeTransactionStatus{}, fmt.Errorf("XTB TradeTransactionStatus processor: %w", err)
	}

	wsTradeTransResp := response.TradeTransactionStatus{}
	if err := json.Unmarshal(resp, &wsTradeTransResp); err != nil {
		return response.TradeTransactionStatus{}, fmt.Errorf("XTB TradeTransactionStatus processor: %w", err)
	}
	return wsTradeTransResp, err
}
