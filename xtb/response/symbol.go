package response

type GetSymbol struct {
	Status     bool `json:"status"`
	ReturnData struct {
		Symbol            string      `json:"symbol"`
		Currency          string      `json:"currency"`
		ContractSize      int         `json:"contractSize"`
		Time              int64       `json:"time"`
		Precision         int         `json:"precision"`
		Type              int         `json:"type"`
		LongOnly          bool        `json:"longOnly"`
		TrailingEnabled   bool        `json:"trailingEnabled"`
		Percentage        float64     `json:"percentage"`
		Bid               float64     `json:"bid"`
		Ask               float64     `json:"ask"`
		High              float64     `json:"high"`
		Low               float64     `json:"low"`
		LotMin            float64     `json:"lotMin"`
		LotMax            float64     `json:"lotMax"`
		LotStep           float64     `json:"lotStep"`
		TickSize          float64     `json:"tickSize"`
		TickValue         float64     `json:"tickValue"`
		SpreadRaw         float64     `json:"spreadRaw"`
		SpreadTable       float64     `json:"spreadTable"`
		Starting          interface{} `json:"starting"`
		MarginMaintenance int         `json:"marginMaintenance"`
		MarginHedged      int         `json:"marginHedged"`
		InitialMargin     int         `json:"initialMargin"`
		TimeString        string      `json:"timeString"`
		ShortSelling      bool        `json:"shortSelling"`
		CurrencyPair      bool        `json:"currencyPair"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
}

type GetSymbolExtended struct {
	Status     bool `json:"status"`
	ReturnData struct {
		Symbol             string      `json:"symbol"`
		Currency           string      `json:"currency"`
		CategoryName       string      `json:"categoryName"`
		CurrencyProfit     string      `json:"currencyProfit"`
		QuoteId            int         `json:"quoteId"`
		QuoteIdCross       int         `json:"quoteIdCross"`
		MarginMode         int         `json:"marginMode"`
		ProfitMode         int         `json:"profitMode"`
		PipsPrecision      int         `json:"pipsPrecision"`
		ContractSize       int         `json:"contractSize"`
		Exemode            int         `json:"exemode"`
		Time               int64       `json:"time"`
		Expiration         interface{} `json:"expiration"`
		StopsLevel         int         `json:"stopsLevel"`
		Precision          int         `json:"precision"`
		SwapType           int         `json:"swapType"`
		StepRuleId         int         `json:"stepRuleId"`
		Type               int         `json:"type"`
		InstantMaxVolume   int         `json:"instantMaxVolume"`
		GroupName          string      `json:"groupName"`
		Description        string      `json:"description"`
		LongOnly           bool        `json:"longOnly"`
		TrailingEnabled    bool        `json:"trailingEnabled"`
		MarginHedgedStrong bool        `json:"marginHedgedStrong"`
		SwapEnable         bool        `json:"swapEnable"`
		Percentage         float64     `json:"percentage"`
		Bid                float64     `json:"bid"`
		Ask                float64     `json:"ask"`
		High               float64     `json:"high"`
		Low                float64     `json:"low"`
		LotMin             float64     `json:"lotMin"`
		LotMax             float64     `json:"lotMax"`
		LotStep            float64     `json:"lotStep"`
		TickSize           float64     `json:"tickSize"`
		TickValue          float64     `json:"tickValue"`
		SwapLong           float64     `json:"swapLong"`
		SwapShort          float64     `json:"swapShort"`
		Leverage           float64     `json:"leverage"`
		SpreadRaw          float64     `json:"spreadRaw"`
		SpreadTable        float64     `json:"spreadTable"`
		Starting           interface{} `json:"starting"`
		SwapRollover3Days  int         `json:"swap_rollover3days"`
		MarginMaintenance  int         `json:"marginMaintenance"`
		MarginHedged       int         `json:"marginHedged"`
		InitialMargin      int         `json:"initialMargin"`
		TimeString         string      `json:"timeString"`
		ShortSelling       bool        `json:"shortSelling"`
		CurrencyPair       bool        `json:"currencyPair"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
}
