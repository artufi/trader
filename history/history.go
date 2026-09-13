package history

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/artufi/trader/config"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/processor"
	"github.com/artufi/trader/xtb/response"
	ws "github.com/gorilla/websocket"
)

const (
	serviceName = "HistoryService"
	// retryBackoff defines the wait duration after failed dialing for a new stream client.
	retryBackoff = 10 * time.Second
)

type orderStore interface {
	SelectOpenOrdersWithPosition() ([]model.Order, error)
	SelectOpenOrders() ([]model.Order, error)
	UpdateClosed(orderID int, details model.OrderClosedDetails) (int, error)
	MarkFailedAsClosed(orderID int) (int, error)
}

type HService struct {
	Cfg    config.AppConfig
	Logger *slog.Logger

	WSManager *websocket.WSManager
	Dialer    *ws.Dialer

	OrderService     orderStore
	OrdersByPosition map[int]int

	sync.Mutex
}

func (hs *HService) GetTradesStream(ctx context.Context, userID string) {
	ctx = logging.AppendAttrsCtx(ctx, logging.ServiceName(serviceName), logging.UserIDAttr(userID))
	for {
		if ctx.Err() != nil {
			hs.Logger.InfoContext(ctx, "Context done, stop dialing for a new stream client to process getTrades",
				logging.ErrorAttr(ctx.Err()))
			return
		}
		wsClientStream, err := hs.WSManager.DialForNewStreamClient(ctx, hs.Cfg.XTB.Demo.WebSocketStreamURL, nil, userID)
		if err != nil {
			hs.Logger.ErrorContext(ctx, "Failed to dial for a new stream client to process getTrades will try again...",
				logging.ErrorAttr(err))
			// TODO: implement retry with a max attempt limit.
			time.Sleep(retryBackoff)
			continue
		}
		ssid := wsClientStream.StreamSessionID
		ssidCtx := logging.AppendAttrsCtx(ctx, logging.StreamID(ssid))

		proc := processor.NewProc(wsClientStream)
		_, err = proc.GetTradesStream(ssidCtx, ssid)
		if err != nil {
			hs.Logger.ErrorContext(ssidCtx, "Failed to start getTrades stream", logging.ErrorAttr(err))
			continue
		}

		if reconnect := hs.readTradesStream(ssidCtx, wsClientStream); !reconnect {
			return
		}
	}
}

func (hs *HService) readTradesStream(ctx context.Context, wsClientStream *websocket.WSClientStream) bool {
	for {
		select {
		case <-ctx.Done():
			hs.Logger.InfoContext(ctx, "Context done, stop getTrades stream read loop")
			return false
		case resp := <-wsClientStream.ReaderRespCh:
			if resp.Err != nil {
				hs.Logger.ErrorContext(ctx, "Failed to read getTrades stream", logging.ErrorAttr(resp.Err))
				return true
			}

			tradeResponseStream := response.GetTradeStream{}
			err := json.Unmarshal(resp.Data, &tradeResponseStream)
			if err != nil {
				hs.Logger.ErrorContext(ctx, "Failed to deserialize getTrades stream data", logging.ErrorAttr(err))
				continue
			}
			hs.Logger.InfoContext(ctx, "getTrades stream response", logging.RespAttr(tradeResponseStream))

			hs.handleTradeUpdate(ctx, tradeResponseStream.Data)
		}
	}
}

// The streaming process of getTrades in XTB is complex in the context of the received response
// because it doesn't work as intuition suggests. For instance, closing a position doesn't result
// in just one response back to the client. Instead, during both opening and closing positions,
// the server returns multiple responses detailing each step.
//
// For example, when opening a position, the server returns several responses informing about
// various stages of processing in XTB. First, there's information about the modified status
// of the created position, likely indicating that a specific operation has been created in their system.
// Then, there's a message about how the position was modified (status modified), presumably indicating
// the inclusion of broker commissions. Next, there's information about the deleted status,
// likely simply approving the previous transaction step indicating acknowledge of the modification
// of this position (presumably recording each step in the database to conduct transactions).
//
// So, all these messages that come through are probably individual stages of transactions executed
// in their system. When opening a position, a trade transaction is initiated for a specific amount,
// which must be adjusted for broker fees. The system informs about every step, including
// the opening of the transaction, subsequent price modifications, final position opening
// with fees considered, etc. This eventually reaches the Deleted state as such transactions is confirmed
// and can be removed from the system.
//
// In the case of closing transactions, once the user-defined goal is reached, such as SL or TP,
// a message is received indicating the position is being closed (Modified state) with specific
// close_price and profit details. Then, this Modified state is removed, and it becomes Deleted
// as confirmation of processed Modified step (prices adjustments). Finally, a message arrives
// again indicating the Modified state, signifying the final closure of the position and providing
// final statistics. Because milliseconds matter in this system, there may be differences even
// in the last stage, and the position is considered Modified again, but now without being Deleted.
// The Deleted state is likely a confirmation, but apparently deemed unnecessary here since
// the position is already closed.
//
// Currently, I haven't found a better way to distinguish unnecessary transactions
// from those I'm looking for because the positionID and orderID are different in
// the case of closing positions. When closing a position, the last message containing
// the most important details has a position number assigned earlier but a different
// orderID (usually, closing positions consist of 3 messages: the first has an orderID
// consistent with the database, the next one has an orderID consistent with the database,
// and the final one has an orderID incremented by 2). Even though the details of this
// operation are related to the sought orderID, it's necessary to save the position as
// a key in the map (because it's always the same), and to save the orderID as a value,
// but remember that in the final step, you shouldn't save the orderID because it will
// be a different value than the orderID saved in the database during position opening.
// It's also important to remove these elements from the map to prevent it from overflowing.
//
// Unfortunately, in this case, a record with a position number the same as the orderID,
// indicating the opening of the position, is also stored in the map, so this unnecessary
// (for now) entry must be removed from the map to prevent overflow. At the moment,
// I haven't found any indicator that specifies that this is an opening position transaction.
// I've only managed to exclude operations like PENDING, which don't contribute anything
// to the system and aren't added to the map.
func (hs *HService) handleTradeUpdate(ctx context.Context, tradeData response.StreamingTradeRecord) {
	position := tradeData.Position
	if !tradeData.Closed && tradeData.Type != int(command.PENDING) {
		hs.Lock()
		hs.OrdersByPosition[position] = tradeData.Order2
		hs.Unlock()
		return
	}
	if !tradeData.Closed || tradeData.Type != int(command.CLOSE) {
		return
	}

	updateDetails := model.OrderClosedDetails{
		ClosePrice: tradeData.ClosePrice,
		OpenPrice:  tradeData.OpenPrice,
		Profit:     tradeData.Profit,
		Closed:     tradeData.Closed,
		Comment:    tradeData.Comment,
		OpenTime:   unixMilli(tradeData.OpenTime),
		CloseTime:  unixMilli(tradeData.CloseTime),
	}

	hs.Lock()
	defer hs.Unlock()

	orderID := hs.OrdersByPosition[position]
	dbID, err := hs.OrderService.UpdateClosed(orderID, updateDetails)
	if err != nil {
		hs.Logger.ErrorContext(ctx, "Failed to update closed order", logging.ErrorAttr(err))
	} else {
		hs.Logger.InfoContext(ctx, "Updated closed order", logging.IDAttr(dbID))
	}

	delete(hs.OrdersByPosition, orderID)
	delete(hs.OrdersByPosition, position)
}

func unixMilli(ms *int) *time.Time {
	if ms == nil {
		return nil
	}
	return new(time.UnixMilli(int64(*ms)))
}
