package processor

import (
	"context"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

const (
	getTradesProc       = "getTrades"
	getTradesStreamProc = "getTradesStream"
)

func (p Proc) GetTrades(ctx context.Context, customTag string, onlyOpened bool) (response.GetTrade, error) {
	getTradesJSON, err := jsonform.GetTrades(customTag, onlyOpened)
	if err != nil {
		return response.GetTrade{}, &ProcError{Processor: getTradesProc, Err: err}
	}

	return process[response.GetTrade](ctx, p.client, customTag, getTradesJSON, getTradesProc)
}

func (p Proc) GetTradesStream(ctx context.Context, sessionID string) ([]byte, error) {
	getTradesJSON, err := jsonform.GetTradesStream(sessionID)
	if err != nil {
		return nil, &ProcError{Processor: getTradesStreamProc, Err: err}
	}

	return processStream(ctx, p.client, "", getTradesJSON, getTradesStreamProc)
}
