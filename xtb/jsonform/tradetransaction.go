package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

func TradeTransaction(tradeTransactionArgs command.TradeTransactionArgs, customTag string) ([]byte, error) {
	tradeTransaction := command.TradeTransaction{
		Command:   "tradeTransaction",
		Arguments: tradeTransactionArgs,
		CustomTag: customTag,
	}
	transactionJSON, err := json.Marshal(tradeTransaction)
	if err != nil {
		return nil, fmt.Errorf("serialize TradeTransaction command: %w", err)
	}
	return transactionJSON, nil
}

func TradeTransactionStatus(orderNo int, customTag string) ([]byte, error) {
	tradeTransactionStatus := command.TradeTransactionStatus{
		Command:   "tradeTransactionStatus",
		CustomTag: customTag,
	}
	tradeTransactionStatus.Arguments.Order = orderNo
	transactionStatusJSON, err := json.Marshal(tradeTransactionStatus)
	if err != nil {
		return nil, fmt.Errorf("serialize TradeTransactionStatus command: %w", err)
	}
	return transactionStatusJSON, nil
}
