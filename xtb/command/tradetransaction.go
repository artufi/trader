package command

type TradeTransaction struct {
	Command   string               `json:"command"`
	Arguments TradeTransactionArgs `json:"arguments"`
}

type TradeTransactionArgs struct {
	TradeTransInfo TradeTransInfo `json:"tradeTransInfo"`
}

type operationCode int

const (
	OPEN operationCode = iota
	PENDING
	CLOSE
	MODIFY
	DELETE
)

type transactionType int

const (
	BUY transactionType = iota
	SELL
	BUY_LIMIT
	SELL_LIMIT
	BUY_STOP
	SELL_STOP
	BALANCE
	CREDIT
)

type TradeTransInfo struct {
	Cmd           operationCode   `json:"cmd"`
	CustomComment string          `json:"customComment"`
	Expiration    int64           `json:"expiration"`
	Offset        int             `json:"offset,omitempty"`
	Order         int             `json:"order,omitempty"`
	Price         float64         `json:"price"`
	Sl            float64         `json:"sl,omitempty"`
	Symbol        string          `json:"symbol"`
	Tp            float64         `json:"tp,omitempty"`
	Type          transactionType `json:"type"`
	Volume        float64         `json:"volume"`
}
