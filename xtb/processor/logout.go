package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"trader/infrastructure/websocket"
	"trader/xtb/jsonform"
	"trader/xtb/response"
)

func Logout(ctx context.Context, wsClient *websocket.WSClient) (response.GeneralResponse, error) {
	logoutJSON, err := jsonform.Logout()
	if err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB Logout command: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, logoutJSON)
	if err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB Logout command: %w", err)
	}

	wsLogoutResp := response.GeneralResponse{}
	if err := json.Unmarshal(resp, &wsLogoutResp); err != nil {
		return response.GeneralResponse{}, fmt.Errorf("failed to execute XTB Logout command: %w", err)
	}

	return wsLogoutResp, err
}
