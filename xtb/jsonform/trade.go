package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

const (
	getTrades = "getTrades"
)

func GetTrades(customTag string, openedOnly bool) ([]byte, error) {
	gt := command.GetTrades{
		Command:   getTrades,
		CustomTag: customTag,
	}
	gt.Arguments.OpenedOnly = openedOnly
	gtJSON, err := json.Marshal(gt)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", getTrades, err)
	}
	return gtJSON, nil
}

func GetTradesStream(ssid string) ([]byte, error) {
	gts := command.GetTradesStream{
		Command:         getTrades,
		StreamSessionId: ssid,
	}
	gtJSON, err := json.Marshal(gts)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", getTrades+"Stream", err)
	}
	return gtJSON, nil
}
