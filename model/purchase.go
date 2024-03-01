package model

import (
	"fmt"
	"math"
	"time"
	"trader/xtb/command"
	"trader/xtb/response"
)

// TODO what about other values?
type PurchaseInstruction struct {
	Symbol     string  `json:"symbol"`
	Date       string  `json:"date"`
	Allocation float64 `json:"allocation"`
	PredsProba float64 `json:"preds_proba"`
	StopLoss   float64 `json:"stop_loss"`
	TakeProfit float64 `json:"take_profit"`
	ModelSetup struct {
		ForecastHorizon int    `json:"forecast_horizon"`
		TargetChange    int    `json:"target_change"`
		ModelType       string `json:"model_type"`
	} `json:"model_setup"`
}

func (pi PurchaseInstruction) PrepareTradeTransInfo(symbolData response.GetSymbolExtended, symbolBaseVolume float64, volumeToBuy float64) (command.TradeTransInfo, error) {
	// TODO what if 0?
	if pi.TakeProfit > 0 {
		return pi.PrepareBUYTradeTransInfo(symbolData, symbolBaseVolume, volumeToBuy)
	}
	return pi.PrepareSELLTradeTransInfo(symbolData, symbolBaseVolume, volumeToBuy)
}

func (pi PurchaseInstruction) PrepareBUYTradeTransInfo(symbolData response.GetSymbolExtended, symbolBaseVolume float64, volumeToBuy float64) (command.TradeTransInfo, error) {
	// precision to two decimal places
	precision := math.Pow(10, 2)
	// calculate the price for the volume you want to buy,
	// taking into account the base amount for a specific instrument - it will not necessarily always be 1!
	volumePrice := (symbolData.ReturnData.Ask * volumeToBuy) / symbolBaseVolume
	tp := math.Round((volumePrice+(volumePrice*pi.TakeProfit/100))*precision) / precision
	sl := math.Round((volumePrice-(volumePrice*pi.StopLoss/100))*precision) / precision

	tradeTransInfo := command.TradeTransInfo{
		Expiration: time.Now().Add(time.Minute * 10).UnixMilli(),
		Price:      symbolData.ReturnData.Ask,
		Symbol:     symbolData.ReturnData.Symbol,
		Type:       command.OPEN,
		Volume:     volumeToBuy,
		Sl:         sl,
		Tp:         tp,
	}
	if pi.PredsProba >= 0.5 {
		tradeTransInfo.Cmd = command.BUY
	} else {
		return command.TradeTransInfo{}, fmt.Errorf("preds_proba low value: %v", pi.PredsProba)
	}
	return tradeTransInfo, nil
}

func (pi PurchaseInstruction) PrepareSELLTradeTransInfo(symbolData response.GetSymbolExtended, symbolBaseVolume float64, volumeToBuy float64) (command.TradeTransInfo, error) {
	// precision to two decimal places
	precision := math.Pow(10, 2)
	// calculate the price for the volume you want to buy,
	// taking into account the base amount for a specific instrument - it will not necessarily always be 1!
	volumePrice := (symbolData.ReturnData.Ask * volumeToBuy) / symbolBaseVolume
	tp := math.Round((volumePrice-(volumePrice*pi.TakeProfit/100))*precision) / precision
	sl := math.Round((volumePrice+(volumePrice*pi.StopLoss/100))*precision) / precision

	tradeTransInfo := command.TradeTransInfo{
		Expiration: time.Now().Add(time.Minute * 10).UnixMilli(),
		Price:      symbolData.ReturnData.Ask,
		Symbol:     symbolData.ReturnData.Symbol,
		Type:       command.OPEN,
		Volume:     volumeToBuy,
		Sl:         sl,
		Tp:         tp,
	}
	if pi.PredsProba >= 0.5 {
		tradeTransInfo.Cmd = command.SELL
	} else {
		return command.TradeTransInfo{}, fmt.Errorf("preds_proba low value: %v", pi.PredsProba)
	}
	return tradeTransInfo, nil
}
