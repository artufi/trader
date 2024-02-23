package json

import (
	"encoding/json"
	"fmt"
	"trader/xtb/command"
)

func TradeTransaction(tradeTransactionArgs command.TradeTransactionArgs) ([]byte, error) {
	tradeTransaction := command.TradeTransaction{
		Command:   "tradeTransaction",
		Arguments: tradeTransactionArgs,
	}
	transactionJSON, err := json.Marshal(tradeTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tradeTransaction data: %w", err)
	}
	return transactionJSON, nil
}
