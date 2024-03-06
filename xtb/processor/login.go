package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

func Login(ctx context.Context, cfg config.AppConfig, wsClient *websocket.WSClient, customTag string) (response.General, error) {
	loginJSON, err := jsonform.Login(
		command.LoginArgs{
			UserID:   wsClient.GetUserID(),
			Password: cfg.Testing.Password,
		},
		customTag)
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
