package jsonform

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/xtb/command"
)

const ping = "ping"

func Ping(customTag string) ([]byte, error) {
	p := command.Ping{
		Command:   ping,
		CustomTag: customTag,
	}
	pJSON, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", ping, err)
	}
	return pJSON, nil
}

func PingStream(ssid string) ([]byte, error) {
	p := command.Ping{
		Command:         ping,
		StreamSessionId: ssid,
	}
	pJSON, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("serialize %s command: %w", ping+"Stream", err)
	}
	return pJSON, nil
}
