package jsonform

import (
	"encoding/json"
	"fmt"
	"trader/xtb/command"
)

func GetSymbol(getSymbolArgs command.GetSymbolArgs) ([]byte, error) {
	getSymbol := command.GetSymbol{
		Command:   "getSymbol",
		Arguments: getSymbolArgs,
	}
	getSymbolJSON, err := json.Marshal(getSymbol)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tradeTransaction data: %w", err)
	}
	return getSymbolJSON, nil
}
