package model

import (
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/response"
	"math"
	"time"
)

type PredictionDetails struct {
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
	RunTimestamp      string `json:"runtimestamp"`
	SlTpLogic         string `json:"sl_tp_logic"`
	Interval          string `json:"interval"`
	IdModelProperties string `json:"id_model_properties"`
}

const (
	Buy      = "Buy"
	Sell     = "Sell"
	NoAction = "NoAction"
)

func (pd PredictionDetails) PrepareTradeTransInfo(symbolData response.GetSymbolExtended, volumeToBuy float64) (command.TradeTransInfo, error) {
	// TODO what if 0?
	switch pd.ModelSetup.ModelType {
	case Buy:
		if pd.TakeProfit >= 0 {
			return pd.prepareBUYTradeTransInfo(symbolData, volumeToBuy)
		}
	case Sell:
		if pd.TakeProfit < 0 {
			return pd.prepareSELLTradeTransInfo(symbolData, volumeToBuy)
		}
	case NoAction:
		return command.TradeTransInfo{}, fmt.Errorf("prepare TradeTransInfo: %q", NoAction)
	}
	return command.TradeTransInfo{}, fmt.Errorf("prepare TradeTransInfo: Unknown ModelType: %q", pd.ModelSetup.ModelType)
}

// TODO
// what about order and offset?
func (pd PredictionDetails) prepareBUYTradeTransInfo(symbolData response.GetSymbolExtended, volumeToBuy float64) (command.TradeTransInfo, error) {
	// precision to two decimal places
	precision := math.Pow(10, 2)

	buyPrice := symbolData.ReturnData.Ask
	tp := math.Round((buyPrice+(buyPrice*math.Abs(pd.TakeProfit/100)))*precision) / precision
	sl := math.Round((buyPrice-(buyPrice*pd.StopLoss/100))*precision) / precision

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
	if pd.PredsProba >= 0.5 {
		tradeTransInfo.Cmd = command.BUY
	} else {
		return command.TradeTransInfo{}, fmt.Errorf("prepare TradeTransInfo: preds_proba low value: %v", pd.PredsProba)
	}
	return tradeTransInfo, nil
}

func (pd PredictionDetails) prepareSELLTradeTransInfo(symbolData response.GetSymbolExtended, volumeToBuy float64) (command.TradeTransInfo, error) {
	// precision to two decimal places
	precision := math.Pow(10, 2)
	sellPrice := symbolData.ReturnData.Bid
	tp := math.Round((sellPrice-(sellPrice*math.Abs(pd.TakeProfit)/100))*precision) / precision
	sl := math.Round((sellPrice+(sellPrice*pd.StopLoss/100))*precision) / precision

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
	if pd.PredsProba >= 0.5 {
		tradeTransInfo.Cmd = command.SELL
	} else {
		return command.TradeTransInfo{}, fmt.Errorf("prepare TradeTransInfo: preds_proba low value: %v", pd.PredsProba)
	}
	return tradeTransInfo, nil
}
