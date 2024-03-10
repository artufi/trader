package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
	"strings"
)

func (p Proc) GetSymbolExtended(ctx context.Context, symbol string, customTag string) (response.GetSymbolExtended, error) {
	symbolJSON, err := jsonform.GetSymbol(
		command.GetSymbolArgs{Symbol: strings.ToUpper(symbol)},
		customTag)
	if err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	resp, err := p.Client.WriteText(ctx, customTag, symbolJSON)
	if err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	wsSymbolResp := response.GetSymbolExtended{}
	if err := json.Unmarshal(resp, &wsSymbolResp); err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	return wsSymbolResp, err
}

func GetSymbolExtended(ctx context.Context, symbol string, wsClient *websocket.WSClient, customTag string) (response.GetSymbolExtended, error) {
	symbolJSON, err := jsonform.GetSymbol(
		command.GetSymbolArgs{Symbol: strings.ToUpper(symbol)},
		customTag)
	if err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, customTag, symbolJSON)
	if err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	wsSymbolResp := response.GetSymbolExtended{}
	if err := json.Unmarshal(resp, &wsSymbolResp); err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	return wsSymbolResp, err
}
