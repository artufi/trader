package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"trader/infrastructure/websocket"
	"trader/xtb/command"
	"trader/xtb/jsonform"
	"trader/xtb/response"
)

func GetSymbol(ctx context.Context, symbol string, wsClient websocket.WSClient) (response.GetSymbolExtended, error) {
	symbolJSON, err := jsonform.GetSymbol(command.GetSymbolArgs{
		Symbol: strings.ToUpper(symbol),
	})
	if err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, symbolJSON)
	if err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	wsSymbolResp := response.GetSymbolExtended{}
	if err := json.Unmarshal(resp, &wsSymbolResp); err != nil {
		return response.GetSymbolExtended{}, fmt.Errorf("XTB GetSymbol procesor: %w", err)
	}

	return wsSymbolResp, err
}
