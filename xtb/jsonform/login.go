package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

const login = "login"

func Login(loginArgs command.LoginArgs, customTag string) ([]byte, error) {
	l := command.Login{
		Command:   login,
		Arguments: loginArgs,
		CustomTag: customTag,
	}
	lJSON, err := json.Marshal(l)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", login, err)
	}
	return lJSON, nil
}
