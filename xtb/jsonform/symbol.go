package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

func GetSymbol(getSymbolArgs command.GetSymbolArgs, customTag string) ([]byte, error) {
	getSymbol := command.GetSymbol{
		Command:   "getSymbol",
		Arguments: getSymbolArgs,
		CustomTag: customTag,
	}
	getSymbolJSON, err := json.Marshal(getSymbol)
	if err != nil {
		return nil, fmt.Errorf("serialize TradeTransaction command: %w", err)
	}
	return getSymbolJSON, nil
}
