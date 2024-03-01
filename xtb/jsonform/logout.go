package jsonform

import (
	"encoding/json"
	"fmt"
	"trader/xtb/command"
)

func Logout() ([]byte, error) {
	logout := command.Logout{
		Command: "logout",
	}
	logoutJSON, err := json.Marshal(logout)
	if err != nil {
		return nil, fmt.Errorf("serialize Logout command: %w", err)
	}
	return logoutJSON, nil
}
