package processor

import (
	"context"
	"errors"
	"github.com/artufi/trader/xtb/response"
	"testing"
)

type MockClient struct {
	fn func(ctx context.Context, id string, data []byte) ([]byte, error)
}

func (m MockClient) WriteText(ctx context.Context, id string, data []byte) ([]byte, error) {
	return m.fn(ctx, id, data)
}

func TestProc_process(t *testing.T) {
	t.Run("err while executing client.WriteText, returned err is wrapped into *ProcError", func(t *testing.T) {
		expectedErr := errors.New("some error inside *ProcError")

		c := MockClient{fn: func(ctx context.Context, id string, data []byte) ([]byte, error) {
			return nil, expectedErr
		}}

		resp, err := process[interface{}](nil, c, "testID", []byte("testData"), "test")
		var pe *ProcError
		if !errors.As(err, &pe) {
			t.Fatalf("expected err to be instance of *ProcError, got: %T", err)
		}
		if err == nil {
			t.Errorf("expected wrapped err: %v, got err: nil", err)
		}
		if resp != nil {
			t.Errorf("expected response: %v, got response: nil", resp)
		}
	})

	t.Run("err when data is not in JSON format", func(t *testing.T) {
		c := MockClient{fn: func(ctx context.Context, id string, data []byte) ([]byte, error) {
			return data, nil
		}}

		wrongJSONData := []byte("not JSON data }")

		resp, err := process[interface{}](nil, c, "testID", wrongJSONData, "test")
		var pe *ProcError
		if !errors.As(err, &pe) {
			t.Fatalf("expected err to be instance of *ProcError, got: %T", err)
		}
		if pe.Err == nil {
			t.Errorf("expected err: %v, got err: nil", pe.Err)
		}
		if resp != nil {
			t.Errorf("expected response: %v, got response: nil", resp)
		}
	})

	t.Run("data has to be in JSON format", func(t *testing.T) {
		c := MockClient{fn: func(ctx context.Context, id string, data []byte) ([]byte, error) {
			return data, nil
		}}

		expectedResponse := response.General{
			Status:    true,
			CustomTag: "123456ResponseCustomTest",
		}
		jsonBytes := []byte(`{"status":true,"customTag":"123456ResponseCustomTest"}`)

		resp, err := process[response.General](nil, c, "testID", jsonBytes, "test")
		if err != nil {
			t.Errorf("expected nil err, got err: %v", err)
		}
		// struct containing map[string]interface{} cannot be compared
		if resp.Status != expectedResponse.Status {
			t.Errorf("expected response: %v, got response: %v", resp, expectedResponse)
		}
		if resp.CustomTag != expectedResponse.CustomTag {
			t.Errorf("expected response: %v, got response: %v", resp, expectedResponse)
		}
	})
}
