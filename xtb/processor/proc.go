package processor

import (
	"context"
	"encoding/json"
	"fmt"
)

type Client interface {
	WriteText(ctx context.Context, id string, data []byte) ([]byte, error)
}

type Proc struct {
	Client Client
}

func process[T any](ctx context.Context, client Client, id string, data []byte, processor string) (T, error) {
	var result T

	respBytes, err := client.WriteText(ctx, id, data)
	if err != nil {
		return result, fmt.Errorf("XTB %s processor: %w", processor, err)
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return result, fmt.Errorf("XTB %s processor: %w", processor, err)
	}

	return result, err
}
