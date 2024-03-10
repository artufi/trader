package processor

import (
	"context"
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
	"strings"
)

const getSymbolExtProc = "GetSymbol"

func (p Proc) GetSymbolExtended(ctx context.Context, symbol string, customTag string) (response.GetSymbolExtended, error) {
	symbolJSON, err := jsonform.GetSymbol(
		command.GetSymbolArgs{Symbol: strings.ToUpper(symbol)},
		customTag)
	if err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB %s procesor: %w", getSymbolExtProc, err)
	}

	return process[response.GetSymbolExtended](ctx, p.Client, customTag, symbolJSON, getSymbolExtProc)
}
