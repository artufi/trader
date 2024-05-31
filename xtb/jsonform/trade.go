package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

const (
	getTrades        = "getTrades"
	getTradesHistory = "getTradesHistory"
	getTradeRecords  = "getTradeRecords"
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

func GetTradesHistory(customTag string, start, end int) ([]byte, error) {
	gth := command.GetTradesHistory{
		Command:   getTradesHistory,
		CustomTag: customTag,
	}
	gth.Arguments.Start = start
	gth.Arguments.End = end
	gthJSON, err := json.Marshal(gth)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", getTradesHistory, err)
	}
	return gthJSON, nil
}

func GetTradeRecords(customTag string, orders []int) ([]byte, error) {
	gtr := command.GetTradeRecords{
		Command:   getTradeRecords,
		CustomTag: customTag,
	}
	gtr.Arguments.Orders = orders
	gtrJSON, err := json.Marshal(gtr)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", getTradeRecords, err)
	}
	return gtrJSON, nil
}
