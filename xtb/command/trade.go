package command

type GetTrades struct {
	Command   string `json:"command"`
	Arguments struct {
		OpenedOnly bool `json:"openedOnly"`
	} `json:"arguments,omitempty"`
	CustomTag string `json:"customTag"`
}

type GetTradesStream struct {
	Command         string `json:"command"`
	StreamSessionId string `json:"streamSessionId"`
}

type GetTradesHistory struct {
	Command   string `json:"command"`
	Arguments struct {
		End   int `json:"end"`
		Start int `json:"start"`
	} `json:"arguments"`
	CustomTag string `json:"customTag"`
}

type GetTradeRecords struct {
	Command   string `json:"command"`
	Arguments struct {
		Orders []int `json:"orders"`
	} `json:"arguments"`
	CustomTag string `json:"customTag"`
}
