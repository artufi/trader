package command

type TradeTransaction struct {
	Command   string               `json:"command"`
	Arguments TradeTransactionArgs `json:"arguments"`
}

type TradeTransactionArgs struct {
	TradeTransInfo TradeTransInfo `json:"tradeTransInfo"`
}

type transactionType int

const (
	OPEN    transactionType = 0
	PENDING transactionType = 1
	CLOSE   transactionType = 2
	MODIFY  transactionType = 3
	DELETE  transactionType = 4
)

type operationCodeCmd int

const (
	BUY        operationCodeCmd = 0
	SELL       operationCodeCmd = 1
	BUY_LIMIT  operationCodeCmd = 2
	SELL_LIMIT operationCodeCmd = 3
	BUY_STOP   operationCodeCmd = 4
	SELL_STOP  operationCodeCmd = 5
	BALANCE    operationCodeCmd = 6
	CREDIT     operationCodeCmd = 7
)

type TradeTransInfo struct {
	Cmd           operationCodeCmd `json:"cmd"`
	CustomComment string           `json:"customComment"`
	Expiration    int64            `json:"expiration"`
	Offset        int              `json:"offset,omitempty"`
	Order         int              `json:"order,omitempty"`
	Price         float64          `json:"price"`
	Sl            float64          `json:"sl,omitempty"`
	Symbol        string           `json:"symbol"`
	Tp            float64          `json:"tp,omitempty"`
	Type          transactionType  `json:"type"`
	Volume        float64          `json:"volume"`
}

type TradeTransactionStatus struct {
	Command   string `json:"command"`
	Arguments struct {
		Order int `json:"order"`
	} `json:"arguments"`
}
