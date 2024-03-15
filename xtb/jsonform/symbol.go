package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

const getSymbol = "getSymbol"

func GetSymbol(getSymbolArgs command.GetSymbolArgs, customTag string) ([]byte, error) {
	gs := command.GetSymbol{
		Command:   getSymbol,
		Arguments: getSymbolArgs,
		CustomTag: customTag,
	}
	gsJSON, err := json.Marshal(gs)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", getSymbol, err)
	}
	return gsJSON, nil
}
