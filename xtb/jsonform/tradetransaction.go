package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

const (
	tradeTransaction       = "tradeTransaction"
	tradeTransactionStatus = "tradeTransactionStatus"
)

func TradeTransaction(tradeTransInfo command.TradeTransInfo, customTag string) ([]byte, error) {
	tt := command.TradeTransaction{
		Command:   tradeTransaction,
		CustomTag: customTag,
	}
	tt.Arguments.TradeTransInfo = tradeTransInfo
	ttJSON, err := json.Marshal(tt)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", tradeTransaction, err)
	}
	return ttJSON, nil
}

func TradeTransactionStatus(orderNo int, customTag string) ([]byte, error) {
	tts := command.TradeTransactionStatus{
		Command:   tradeTransactionStatus,
		CustomTag: customTag,
	}
	tts.Arguments.Order = orderNo
	ttsJSON, err := json.Marshal(tts)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", tradeTransactionStatus, err)
	}
	return ttsJSON, nil
}
