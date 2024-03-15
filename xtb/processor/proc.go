package processor

import (
	"context"
	"encoding/json"
	"fmt"
)

type Client interface {
	WriteText(ctx context.Context, id string, data []byte) ([]byte, error)
}

func NewProc(client Client) Proc {
	return Proc{client: client}
}

type Proc struct {
	client Client
}

type ProcError struct {
	Processor string
	Err       error
}

func (p *ProcError) Error() string {
	return fmt.Sprintf("processor %s: %v", p.Processor, p.Err)
}

func process[T any](ctx context.Context, client Client, id string, data []byte, processor string) (T, error) {
	var result T

	respBytes, err := client.WriteText(ctx, id, data)
	if err != nil {
		return result, &ProcError{Processor: processor, Err: err}
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return result, &ProcError{Processor: processor, Err: err}
	}

	return result, err
}
