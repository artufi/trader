package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

const logout = "logout"

func Logout(customTag string) ([]byte, error) {
	l := command.Logout{
		Command:   logout,
		CustomTag: customTag,
	}
	lJSON, err := json.Marshal(l)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", logout, err)
	}
	return lJSON, nil
}
