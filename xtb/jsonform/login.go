package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

func Login(loginArgs command.LoginArgs, customTag string) ([]byte, error) {
	login := command.Login{
		Command:   "login",
		Arguments: loginArgs,
		CustomTag: customTag,
	}
	loginJSON, err := json.Marshal(login)
	if err != nil {
		return nil, fmt.Errorf("serialize Login command: %w", err)
	}
	return loginJSON, nil
}
