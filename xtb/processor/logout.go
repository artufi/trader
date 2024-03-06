package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

func Logout(ctx context.Context, wsClient *websocket.WSClient, customTag string) (response.General, error) {
	logoutJSON, err := jsonform.Logout(customTag)
	if err != nil {
		return response.General{}, fmt.Errorf("XTB Logout processor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, logoutJSON)
	if err != nil {
		return response.General{}, fmt.Errorf("XTB Logout processor: %w", err)
	}

	wsLogoutResp := response.General{}
	if err := json.Unmarshal(resp, &wsLogoutResp); err != nil {
		return response.General{}, fmt.Errorf("XTB Logout processor: %w", err)
	}

	return wsLogoutResp, err
}
