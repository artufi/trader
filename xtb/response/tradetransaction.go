package response

type TradeTransaction struct {
	Status          bool   `json:"status"`
	StreamSessionId string `json:"streamSessionId"`
	ReturnData      struct {
		Order int `json:"order"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
}

type TradeTransactionStatus struct {
	Status          bool   `json:"status"`
	StreamSessionId string `json:"streamSessionId"`
	ReturnData      struct {
		Ask           float64     `json:"ask"`
		Bid           float64     `json:"bid"`
		CustomComment string      `json:"customComment"`
		Message       interface{} `json:"message"` // not sure maybe pointer?
		Order         int         `json:"order"`   // not sure maybe float64
		RequestStatus int         `json:"requestStatus"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
}
