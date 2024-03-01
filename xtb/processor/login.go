package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"trader/config"
	"trader/infrastructure/websocket"
	"trader/xtb/command"
	"trader/xtb/jsonform"
	"trader/xtb/response"
)

func Login(ctx context.Context, cfg config.AppConfig, wsClient *websocket.WSClient) (response.General, error) {
	loginJSON, err := jsonform.Login(command.LoginArgs{
		UserID:   wsClient.GetUserID(),
		Password: cfg.Testing.Password,
	})
	if err != nil {
		return response.General{}, fmt.Errorf("XTB Login processor: %w", err)
	}

	resp, err := wsClient.WriteText(ctx, loginJSON)
	if err != nil {
		return response.General{}, fmt.Errorf("XTB Login processor: %w", err)
	}

	wsLoginResp := response.General{}
	if err := json.Unmarshal(resp, &wsLoginResp); err != nil {
		return response.General{}, fmt.Errorf("XTB Login processor: %w", err)
	}

	return wsLoginResp, err
}
