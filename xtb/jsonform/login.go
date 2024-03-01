package jsonform

import (
	"encoding/json"
	"fmt"
	"trader/xtb/command"
)

func Login(loginArgs command.LoginArgs) ([]byte, error) {
	login := command.Login{
		Command:   "login",
		Arguments: loginArgs,
	}
	loginJSON, err := json.Marshal(login)
	if err != nil {
		return nil, fmt.Errorf("serialize Login command: %w", err)
	}
	return loginJSON, nil
}
