package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

func Logout(customTag string) ([]byte, error) {
	logout := command.Logout{
		Command:   "logout",
		CustomTag: customTag,
	}
	logoutJSON, err := json.Marshal(logout)
	if err != nil {
		return nil, fmt.Errorf("serialize Logout command: %w", err)
	}
	return logoutJSON, nil
}
