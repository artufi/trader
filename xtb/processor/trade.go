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

func (p Proc) GetTrades(ctx context.Context, customTag string) (response.TradeResponse, error) {
	getTradesJSON, err := jsonform.GetTrades(customTag)
	if err != nil {
		return response.TradeResponse{}, &ProcError{Processor: getTradesProc, Err: err}
	}

	return process[response.TradeResponse](ctx, p.client, customTag, getTradesJSON, getTradesProc)
}

func (p Proc) GetTradesStream(ctx context.Context, sessionID string) (response.TradeResponseStream, error) {
	getTradesJSON, err := jsonform.GetTradesStream(sessionID)
	if err != nil {
		return response.TradeResponseStream{}, &ProcError{Processor: getTradesStreamProc, Err: err}
	}

	return process[response.TradeResponseStream](ctx, p.client, "", getTradesJSON, getTradesStreamProc)
}
