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
		return nil, fmt.Errorf("failed to serialize login data: %w", err)
	}
	return loginJSON, nil
}
