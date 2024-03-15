package processor

import (
	"context"
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
		return response.TradeTransaction{}, &ProcError{Processor: tradeTransProc, Err: err}
	}

	return process[response.TradeTransaction](ctx, p.client, customTag, tradeTransactionJSON, tradeTransProc)
}

func (p Proc) TradeTransactionStatus(ctx context.Context, orderNo int, customTag string) (response.TradeTransactionStatus, error) {
	tradeTransactionStatusJSON, err := jsonform.TradeTransactionStatus(orderNo, customTag)
	if err != nil {
		return response.TradeTransactionStatus{}, &ProcError{Processor: tradeTransStatusProc, Err: err}
	}

	return process[response.TradeTransactionStatus](ctx, p.client, customTag, tradeTransactionStatusJSON, tradeTransStatusProc)
}
