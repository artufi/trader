package model

import (
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/response"
	"math"
	"time"
)

type PurchaseInstruction struct {
	Symbol         string  `json:"symbol"`
	PredictionDate string  `json:"prediction_date"`
	Allocation     float64 `json:"allocation"`
	PredsProba     float64 `json:"preds_proba"`
	StopLoss       float64 `json:"stop_loss"`
	TakeProfit     float64 `json:"take_profit"`
	ModelSetup     struct {
		ForecastHorizon int    `json:"forecast_horizon"`
		TargetChange    int    `json:"target_change"`
		ModelType       string `json:"model_type"`
	} `json:"model_setup"`
}

const (
	BUY  = "Buy"
	SELL = "Sell"
)

func (pi PurchaseInstruction) PrepareTradeTransInfo(symbolData response.GetSymbolExtended, volumeToBuy float64) (command.TradeTransInfo, error) {
	// TODO what if 0?
	switch pi.ModelSetup.ModelType {
	case BUY:
		if pi.TakeProfit >= 0 {
			return pi.prepareBUYTradeTransInfo(symbolData, volumeToBuy)
		}
	case SELL:
		if pi.TakeProfit < 0 {
			return pi.prepareSELLTradeTransInfo(symbolData, volumeToBuy)
		}
	}
	return command.TradeTransInfo{}, fmt.Errorf("prepare TradeTransInfo: UNKNOWN ModelType")
}

func (pi PurchaseInstruction) prepareBUYTradeTransInfo(symbolData response.GetSymbolExtended, volumeToBuy float64) (command.TradeTransInfo, error) {
	// precision to two decimal places
	precision := math.Pow(10, 2)

	buyPrice := symbolData.ReturnData.Ask
	tp := math.Round((buyPrice+(buyPrice*math.Abs(pi.TakeProfit/100)))*precision) / precision
	sl := math.Round((buyPrice-(buyPrice*pi.StopLoss/100))*precision) / precision

	tradeTransInfo := command.TradeTransInfo{
		CustomComment: "BUY TRANSACTION",
		Expiration:    time.Now().Add(time.Minute * 10).UnixMilli(),
		Price:         buyPrice,
		Symbol:        symbolData.ReturnData.Symbol,
		Type:          command.OPEN,
		Volume:        volumeToBuy,
		Sl:            sl,
		Tp:            tp,
	}
	if pi.PredsProba >= 0.5 {
		tradeTransInfo.Cmd = command.BUY
	} else {
		return command.TradeTransInfo{}, fmt.Errorf("prepare TradeTransInfo: preds_proba low value: %v", pi.PredsProba)
	}
	return tradeTransInfo, nil
}

func (pi PurchaseInstruction) prepareSELLTradeTransInfo(symbolData response.GetSymbolExtended, volumeToBuy float64) (command.TradeTransInfo, error) {
	// precision to two decimal places
	precision := math.Pow(10, 2)
	sellPrice := symbolData.ReturnData.Bid
	tp := math.Round((sellPrice-(sellPrice*math.Abs(pi.TakeProfit)/100))*precision) / precision
	sl := math.Round((sellPrice+(sellPrice*pi.StopLoss/100))*precision) / precision

	tradeTransInfo := command.TradeTransInfo{
		CustomComment: "SELL TRANSACTION",
		Expiration:    time.Now().Add(time.Minute * 10).UnixMilli(),
		Price:         sellPrice,
		Symbol:        symbolData.ReturnData.Symbol,
		Type:          command.OPEN,
		Volume:        volumeToBuy,
		Sl:            sl,
		Tp:            tp,
	}
	if pi.PredsProba >= 0.5 {
		tradeTransInfo.Cmd = command.SELL
	} else {
		return command.TradeTransInfo{}, fmt.Errorf("prepare TradeTransInfo: preds_proba low value: %v", pi.PredsProba)
	}
	return tradeTransInfo, nil
}
