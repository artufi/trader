package processor

import (
	"context"
)

type Client interface {
	WriteText(ctx context.Context, id string, data []byte) ([]byte, error)
	GetUserID() string
}

type Proc struct {
	Client Client
}

//type StatusChecker interface {
//	CheckStatus() error
//}

//func process(ctx context.Context, proc Proc, id string, jsonData []byte, respType StatusChecker, processor string) (StatusChecker, error) {
//	respBytes, err := proc.Client.WriteText(ctx, id, jsonData)
//	if err != nil {
//		return respType, fmt.Errorf("XTB %s processor: %w", processor, err)
//	}
//
//	if err := json.Unmarshal(respBytes, &respType); err != nil {
//		return respType, fmt.Errorf("XTB %s processor: %w", processor, err)
//	}
//
//	return respType, err
//}
