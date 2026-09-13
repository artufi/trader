package history

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/model"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/response"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

type stubOrderStore struct {
	closedOrderID      int
	closedDetails      model.OrderClosedDetails
	updateClosedCalled bool
	updateClosedErr    error
}

func (s *stubOrderStore) SelectOpenOrdersWithPosition() ([]model.Order, error) { return nil, nil }
func (s *stubOrderStore) SelectOpenOrders() ([]model.Order, error)             { return nil, nil }
func (s *stubOrderStore) MarkFailedAsClosed(int) (int, error)                  { return 0, nil }

func (s *stubOrderStore) UpdateClosed(orderID int, details model.OrderClosedDetails) (int, error) {
	s.updateClosedCalled = true
	s.closedOrderID = orderID
	s.closedDetails = details
	return orderID, s.updateClosedErr
}

func newTestHService(store orderStore) *HService {
	return &HService{
		Logger:           newDiscardLogger(),
		OrderService:     store,
		OrdersByPosition: make(map[int]int),
	}
}

func TestHService_handleTradeUpdate(t *testing.T) {
	t.Run("records the position while a not-yet-closed transaction updates its order number", func(t *testing.T) {
		store := &stubOrderStore{}
		hs := newTestHService(store)

		hs.handleTradeUpdate(context.Background(), response.StreamingTradeRecord{
			Position: 100,
			Order2:   200,
			Closed:   false,
			Type:     int(command.OPEN),
		})

		if got := hs.OrdersByPosition[100]; got != 200 {
			t.Errorf("expected order 200 recorded for position 100, got: %d", got)
		}
		if store.updateClosedCalled {
			t.Error("expected no database update for a transaction that is not closed")
		}
	})

	t.Run("skips a PENDING transaction", func(t *testing.T) {
		store := &stubOrderStore{}
		hs := newTestHService(store)

		hs.handleTradeUpdate(context.Background(), response.StreamingTradeRecord{
			Position: 100,
			Order2:   200,
			Closed:   false,
			Type:     int(command.PENDING),
		})

		if _, ok := hs.OrdersByPosition[100]; ok {
			t.Errorf("expected position 100 not to be recorded, got: %v", hs.OrdersByPosition)
		}
	})

	t.Run("updates and forgets the order on the final CLOSE message", func(t *testing.T) {
		store := &stubOrderStore{}
		hs := newTestHService(store)
		hs.OrdersByPosition[100] = 200

		profit := 12.5
		hs.handleTradeUpdate(context.Background(), response.StreamingTradeRecord{
			Position: 100,
			Order2:   999,
			Closed:   true,
			Type:     int(command.CLOSE),
			Profit:   &profit,
		})

		if !store.updateClosedCalled {
			t.Fatal("expected the database update to run")
		}
		if store.closedOrderID != 200 {
			t.Errorf("expected the order recorded when the position opened (200), got: %d", store.closedOrderID)
		}
		if store.closedDetails.Profit == nil || *store.closedDetails.Profit != profit {
			t.Errorf("expected profit: %v, got: %v", profit, store.closedDetails.Profit)
		}
		if _, ok := hs.OrdersByPosition[100]; ok {
			t.Errorf("expected position 100 to be forgotten, got: %v", hs.OrdersByPosition)
		}
		if _, ok := hs.OrdersByPosition[200]; ok {
			t.Errorf("expected order 200 to be forgotten, got: %v", hs.OrdersByPosition)
		}
	})

	t.Run("ignores a closed transaction that is not the CLOSE message", func(t *testing.T) {
		store := &stubOrderStore{}
		hs := newTestHService(store)
		hs.OrdersByPosition[100] = 200

		hs.handleTradeUpdate(context.Background(), response.StreamingTradeRecord{
			Position: 100,
			Closed:   true,
			Type:     int(command.MODIFY),
		})

		if store.updateClosedCalled {
			t.Error("expected no database update")
		}
		if got := hs.OrdersByPosition[100]; got != 200 {
			t.Errorf("expected position 100 to still map to 200, got: %d", got)
		}
	})

	t.Run("converts optional millisecond timestamps, leaving them nil when absent", func(t *testing.T) {
		store := &stubOrderStore{}
		hs := newTestHService(store)
		hs.OrdersByPosition[100] = 200

		hs.handleTradeUpdate(context.Background(), response.StreamingTradeRecord{
			Position: 100,
			Closed:   true,
			Type:     int(command.CLOSE),
		})

		if store.closedDetails.OpenTime != nil {
			t.Errorf("expected a nil open time, got: %v", store.closedDetails.OpenTime)
		}
		if store.closedDetails.CloseTime != nil {
			t.Errorf("expected a nil close time, got: %v", store.closedDetails.CloseTime)
		}

		openMillis := 1_700_000_000_000
		closeMillis := 1_700_000_060_000
		hs.OrdersByPosition[100] = 200
		hs.handleTradeUpdate(context.Background(), response.StreamingTradeRecord{
			Position:  100,
			Closed:    true,
			Type:      int(command.CLOSE),
			OpenTime:  &openMillis,
			CloseTime: &closeMillis,
		})

		if store.closedDetails.OpenTime == nil || !store.closedDetails.OpenTime.Equal(time.UnixMilli(int64(openMillis))) {
			t.Errorf("expected open time: %v, got: %v", time.UnixMilli(int64(openMillis)), store.closedDetails.OpenTime)
		}
		if store.closedDetails.CloseTime == nil || !store.closedDetails.CloseTime.Equal(time.UnixMilli(int64(closeMillis))) {
			t.Errorf("expected close time: %v, got: %v", time.UnixMilli(int64(closeMillis)), store.closedDetails.CloseTime)
		}
	})
}

func TestHService_readTradesStream(t *testing.T) {
	newStreamClient := func() *websocket.WSClientStream {
		return &websocket.WSClientStream{ReaderRespCh: make(chan websocket.ReaderResp, 1)}
	}

	t.Run("reconnects when the reader reports an error", func(t *testing.T) {
		hs := newTestHService(&stubOrderStore{})
		client := newStreamClient()
		client.ReaderRespCh <- websocket.ReaderResp{Err: errors.New("connection reset")}

		reconnect := hs.readTradesStream(context.Background(), client)

		if !reconnect {
			t.Error("expected readTradesStream to report true (reconnect) after a reader error")
		}
	})

	t.Run("stops without reconnecting when the context is done", func(t *testing.T) {
		hs := newTestHService(&stubOrderStore{})
		client := newStreamClient()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		reconnect := hs.readTradesStream(ctx, client)

		if reconnect {
			t.Error("expected readTradesStream to report false (no reconnect) once the context is done")
		}
	})

	t.Run("skips a malformed message and keeps reading", func(t *testing.T) {
		hs := newTestHService(&stubOrderStore{})
		client := newStreamClient()

		client.ReaderRespCh <- websocket.ReaderResp{Data: []byte("not json")}
		go func() {
			client.ReaderRespCh <- websocket.ReaderResp{
				Data: []byte(`{"data":{"position":100,"order2":200,"closed":false,"type":0}}`),
			}
			client.ReaderRespCh <- websocket.ReaderResp{Err: errors.New("connection reset")}
		}()

		reconnect := hs.readTradesStream(context.Background(), client)

		if !reconnect {
			t.Error("expected readTradesStream to report true (reconnect) after the reader error")
		}
		if got := hs.OrdersByPosition[100]; got != 200 {
			t.Errorf("expected the well-formed message to still be processed, got: %v", hs.OrdersByPosition)
		}
	})
}
