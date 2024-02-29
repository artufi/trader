package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"
	"trader/infrastructure/websocket"
	"trader/model"
	"trader/xtb/command"
	"trader/xtb/jsonform"
	"trader/xtb/response"
)

func PrepareTradeTransInfo(tradeInstr model.PurchaseInstruction, symbolData response.GetSymbolExtended) (command.TradeTransInfo, error) {
	pow := math.Pow(10, 2)
	// czemu nie ma tego w odpowiedzi???!!
	cenaZa02wiecTrzebaRazy5Volumen := symbolData.ReturnData.Ask * 5 * 0.2
	tp := math.Round((cenaZa02wiecTrzebaRazy5Volumen+(cenaZa02wiecTrzebaRazy5Volumen*tradeInstr.TakeProfit/100))*pow) / pow
	sl := math.Round((cenaZa02wiecTrzebaRazy5Volumen-(cenaZa02wiecTrzebaRazy5Volumen*tradeInstr.StopLoss/100))*pow) / pow
	tradeTransInfo := command.TradeTransInfo{
		CustomComment: "test",
		Expiration:    time.Now().Add(time.Minute * 10).UnixMilli(),
		Price:         symbolData.ReturnData.Ask,
		Symbol:        symbolData.ReturnData.Symbol,
		Type:          command.BUY,
		Volume:        0.2,
		Sl:            sl,
		// precision!
		Tp: tp,
	}
	if tradeInstr.PredsProba >= 0.5 {
		tradeTransInfo.Cmd = command.OPEN
	} else {
		return command.TradeTransInfo{}, fmt.Errorf("preds_proba too low value: %v", tradeInstr.PredsProba)
	}
	return tradeTransInfo, nil
}

func TradeTransaction(ctx context.Context, tradeTransInfo command.TradeTransInfo, wsClient *websocket.WSClient) (response.GeneralResponse, error) {
	tradeTransactionJSON, err := jsonform.TradeTransaction(command.TradeTransactionArgs{
		TradeTransInfo: tradeTransInfo})
	if err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB TradeTransaction command: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, tradeTransactionJSON)
	if err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB TradeTransaction command: %w", err)
	}

	wsTradeTransResp := response.GeneralResponse{}
	if err := json.Unmarshal(resp, &wsTradeTransResp); err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB TradeTransaction command: %w", err)
	}
	return wsTradeTransResp, err
}

func TradeTransactionStatus(ctx context.Context, orderNo int, wsClient *websocket.WSClient) (response.GeneralResponse, error) {
	tradeTransactionStatusJSON, err := jsonform.TradeTransactionStatus(orderNo)
	if err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB TradeTransactionStatus command: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, tradeTransactionStatusJSON)
	if err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB TradeTransactionStatus command: %w", err)
	}

	wsTradeTransResp := response.GeneralResponse{}
	if err := json.Unmarshal(resp, &wsTradeTransResp); err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB TradeTransaction command: %w", err)
	}
	return wsTradeTransResp, err
}
