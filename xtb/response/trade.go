package response

type GetTrade struct {
	Status     bool          `json:"status"`
	ReturnData []TradeRecord `json:"returnData"`
	ErrorCode  string        `json:"errorCode,omitempty"`
	ErrorDescr string        `json:"errorDescr,omitempty"`
	CustomTag  string        `json:"customTag,omitempty"`
}

type TradeRecord struct {
	ClosePrice       float64     `json:"close_price"`
	CloseTime        interface{} `json:"close_time"`
	CloseTimeString  interface{} `json:"close_timeString"`
	Closed           bool        `json:"closed"`
	Cmd              int         `json:"cmd"`
	Comment          string      `json:"comment"`
	Commission       float64     `json:"commission"`
	CustomComment    string      `json:"customComment"`
	Digits           int         `json:"digits"`
	Expiration       interface{} `json:"expiration"`
	ExpirationString interface{} `json:"expirationString"`
	MarginRate       float64     `json:"margin_rate"`
	Offset           int         `json:"offset"`
	OpenPrice        float64     `json:"open_price"`
	OpenTime         int64       `json:"open_time"`
	OpenTimeString   string      `json:"open_timeString"`
	Order            int         `json:"order"`
	Order2           int         `json:"order2"`
	Position         int         `json:"position"`
	Profit           float64     `json:"profit"`
	Sl               float64     `json:"sl"`
	Storage          float64     `json:"storage"`
	Symbol           string      `json:"symbol"`
	Timestamp        int64       `json:"timestamp"`
	Tp               float64     `json:"tp"`
	Volume           float64     `json:"volume"`
}

type GetTradeStream struct {
	Command    string               `json:"command"`
	Data       StreamingTradeRecord `json:"data"`
	ErrorCode  string               `json:"errorCode,omitempty"`
	ErrorDescr string               `json:"errorDescr,omitempty"`
	CustomTag  string               `json:"customTag,omitempty"`
}

type StreamingTradeRecord struct {
	ClosePrice    float64     `json:"close_price"`
	CloseTime     interface{} `json:"close_time"`
	Closed        bool        `json:"closed"`
	Cmd           int         `json:"cmd"`
	Comment       string      `json:"comment"`
	Commission    float64     `json:"commission"`
	CustomComment string      `json:"customComment"`
	Digits        int         `json:"digits"`
	Expiration    interface{} `json:"expiration"`
	MarginRate    float64     `json:"margin_rate"`
	Offset        int         `json:"offset"`
	OpenPrice     float64     `json:"open_price"`
	OpenTime      int64       `json:"open_time"`
	Order         int         `json:"order"`
	Order2        int         `json:"order2"`
	Position      int         `json:"position"`
	Profit        float64     `json:"profit"`
	Sl            float64     `json:"sl"`
	State         string      `json:"state"`
	Storage       float64     `json:"storage"`
	Symbol        string      `json:"symbol"`
	Tp            float64     `json:"tp"`
	Type          int         `json:"type"`
	Volume        float64     `json:"volume"`
}
