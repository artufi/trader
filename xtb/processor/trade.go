package processor

import (
	"context"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

const (
	getTradesProc        = "getTrades"
	getTradesStreamProc  = "getTradesStream"
	getTradesHistoryProc = "getTradesHistory"
	getTradeRecordsProc  = "getTradeRecords"
)

func (p Proc) GetTrades(ctx context.Context, customTag string, onlyOpened bool) (response.GetTrades, error) {
	getTradesJSON, err := jsonform.GetTrades(customTag, onlyOpened)
	if err != nil {
		return response.GetTrades{}, &ProcError{Processor: getTradesProc, Err: err}
	}

	return process[response.GetTrades](ctx, p.client, customTag, getTradesJSON, getTradesProc)
}

func (p Proc) GetTradesStream(ctx context.Context, sessionID string) ([]byte, error) {
	getTradesJSON, err := jsonform.GetTradesStream(sessionID)
	if err != nil {
		return nil, &ProcError{Processor: getTradesStreamProc, Err: err}
	}

	return processStream(ctx, p.client, "", getTradesJSON, getTradesStreamProc)
}

func (p Proc) GetTradesHistory(ctx context.Context, customTag string, start, end int) (response.GetTrades, error) {
	getTradesHistory, err := jsonform.GetTradesHistory(customTag, start, end)
	if err != nil {
		return response.GetTrades{}, &ProcError{Processor: getTradesHistoryProc, Err: err}
	}

	return process[response.GetTrades](ctx, p.client, customTag, getTradesHistory, getTradesHistoryProc)
}

func (p Proc) GetTradeRecords(ctx context.Context, customTag string, orders []int) (response.GetTrades, error) {
	getTradeRecordsJSON, err := jsonform.GetTradeRecords(customTag, orders)
	if err != nil {
		return response.GetTrades{}, &ProcError{Processor: getTradeRecordsProc, Err: err}
	}

	return process[response.GetTrades](ctx, p.client, customTag, getTradeRecordsJSON, getTradeRecordsProc)
}
