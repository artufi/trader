package model

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
