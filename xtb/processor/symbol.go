package processor

import (
	"context"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
	"strings"
)

const getSymbolExtProc = "getSymbol"

func (p Proc) GetSymbolExtended(ctx context.Context, symbol string, customTag string) (response.GetSymbolExtended, error) {
	symbolJSON, err := jsonform.GetSymbol(
		command.GetSymbolArgs{Symbol: strings.ToUpper(symbol)},
		customTag)
	if err != nil {
		return response.GetSymbolExtended{}, &ProcError{Processor: getSymbolExtProc, Err: err}
	}

	return process[response.GetSymbolExtended](ctx, p.client, customTag, symbolJSON, getSymbolExtProc)
}
