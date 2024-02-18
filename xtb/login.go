package xtb

import (
	"encoding/json"
	"fmt"
	"trader/xtb/command"
)

func Login(loginArgs command.LoginArgs) ([]byte, error) {
	loginData := command.Login{
		Command:   "login",
		Arguments: loginArgs,
	}
	loginJSON, err := json.Marshal(loginData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize login request: %w", err)
	}
	return loginJSON, nil
}
