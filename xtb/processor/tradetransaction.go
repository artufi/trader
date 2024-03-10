package processor

import (
	"context"
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

const (
	tradeTransProc       = "TradeTransaction"
	tradeTransStatusProc = "TradeTransactionStatus"
)

func (p Proc) TradeTransaction(ctx context.Context, tradeTransInfo command.TradeTransInfo, customTag string) (response.TradeTransaction, error) {
	tradeTransactionJSON, err := jsonform.TradeTransaction(tradeTransInfo, customTag)
	if err != nil {
		return response.TradeTransaction{}, fmt.Errorf("XTB %s processor: %w", tradeTransProc, err)
	}

	return process[response.TradeTransaction](ctx, p.Client, customTag, tradeTransactionJSON, tradeTransProc)
}

func (p Proc) TradeTransactionStatus(ctx context.Context, orderNo int, customTag string) (response.TradeTransactionStatus, error) {
	tradeTransactionStatusJSON, err := jsonform.TradeTransactionStatus(orderNo, customTag)
	if err != nil {
		return response.TradeTransactionStatus{}, fmt.Errorf("XTB %s processor: %w", tradeTransStatusProc, err)
	}

	return process[response.TradeTransactionStatus](ctx, p.Client, customTag, tradeTransactionStatusJSON, tradeTransStatusProc)
}
