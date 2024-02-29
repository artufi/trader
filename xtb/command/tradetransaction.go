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
	OPEN    operationCode = 0
	PENDING               = 1
	CLOSE                 = 2
	MODIFY                = 3
	DELETE                = 4
)

type transactionType int

const (
	BUY        transactionType = 0
	SELL                       = 1
	BUY_LIMIT                  = 2
	SELL_LIMIT                 = 3
	BUY_STOP                   = 4
	SELL_STOP                  = 5
	BALANCE                    = 6
	CREDIT                     = 7
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

type TradeTransactionStatus struct {
	Command   string `json:"command"`
	Arguments struct {
		Order int `json:"order"`
	} `json:"arguments"`
}
