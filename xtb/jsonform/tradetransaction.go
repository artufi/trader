package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

func TradeTransaction(tradeTransactionArgs command.TradeTransactionArgs) ([]byte, error) {
	tradeTransaction := command.TradeTransaction{
		Command:   "tradeTransaction",
		Arguments: tradeTransactionArgs,
	}
	transactionJSON, err := json.Marshal(tradeTransaction)
	if err != nil {
		return nil, fmt.Errorf("serialize TradeTransaction command: %w", err)
	}
	return transactionJSON, nil
}

func TradeTransactionStatus(orderNo int) ([]byte, error) {
	tradeTransactionStatus := command.TradeTransactionStatus{
		Command: "tradeTransactionStatus",
	}
	tradeTransactionStatus.Arguments.Order = orderNo
	transactionStatusJSON, err := json.Marshal(tradeTransactionStatus)
	if err != nil {
		return nil, fmt.Errorf("serialize TradeTransactionStatus command: %w", err)
	}
	return transactionStatusJSON, nil
}
