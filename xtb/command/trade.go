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
	StreamSessionId string `json:"streamSessionId,omitempty"`
}
