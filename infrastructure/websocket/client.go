package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/logging"
	"github.com/gorilla/websocket"
	"log/slog"
	"sync"
	"time"
)

const readTimeout = 10

type WSClient struct {
	conn    *websocket.Conn
	manager *WSManager
	logger  *slog.Logger

	// User should send requests in 200 ms intervals.
	// This rule can be broken, but if it happens 6 times in a row the connection is dropped.
	sendRateLimiter *time.Ticker

	// WebSocket connection is only allowed to have one concurrent writer,
	// mutex acts also as a locker for pending map during put and delete operations.
	mutex   sync.Mutex
	pending map[string]*call

	userID string
}

func (wsc *WSClient) GetUserID() string {
	return wsc.userID
}

func (wsc *WSClient) CloseConnection() {
	wsc.conn.Close()
}

func (wsc *WSClient) StopSendRateLimiter() {
	wsc.sendRateLimiter.Stop()
}

func (wsc *WSClient) ReadMessages(ctx context.Context) {
	defer func() {
		wsc.manager.RemoveClient(ctx, wsc)

		// TODO
		// can also add error to call struct to return error cause from reader
		// close and delete call from pending map
		wsc.mutex.Lock()
		for _, c := range wsc.pending {
			close(c.Done)
			delete(wsc.pending, c.ID)
		}
		wsc.mutex.Unlock()
	}()

	chanBreaker := time.NewTimer(time.Second * readTimeout)
	for {
		select {
		// *http.Request context is done after processing request
		// it can be any passed context!
		case <-ctx.Done():
			wsc.logger.InfoContext(ctx, "Context done - reader operation canceled", logging.ErrorAttr(ctx.Err()))
			return
		// TODO
		// conn.ReadMessage is blocking so default case here is not the best pick
		// but context can be done before Read, so it makes sense
		default:
			// blocking method
			_, response, err := wsc.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					wsc.logger.ErrorContext(ctx, "Failed to read a message", logging.ErrorAttr(err))
				}
				// For example after Logout: websocket: close 1006 (abnormal closure): unexpected EOF
				// Logout is normally processed, response (ReadMessage) is read correctly, and after
				// processing Logout conn (connection) is closed with the close 1006 message.
				// This situation also occur after too many Writes to XTB by client.
				wsc.logger.InfoContext(ctx, "Closing reader", logging.ErrorAttr(err))
				return
			}

			// If the Stop method returns false, it indicates that the timer has either
			// expired or was stopped in a previous loop iteration. To enable the reuse
			// of the same timer, it's necessary to read the value from the timer's channel.
			if !chanBreaker.Stop() {
				select {
				case <-chanBreaker.C:
				default:
				}
			}

			// TODO
			// There is a minimal chance that the Stop method may not detect that the timer
			// has expired or was stopped because it was called just before the timer's expiration.
			// Consequently, Reset may return false, and the timer will not be cleared.
			// However, this is super rare.
			// ---
			// This step prevents the immediate reading from the timer's channel and
			// allows for the timer to be reset for the next Read operation.
			chanBreaker.Reset(time.Second * readTimeout)

			// read customTag from response to link to request
			customTag := struct {
				ID string `json:"customTag"`
			}{}
			err = json.Unmarshal(response, &customTag)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Reader unable to unmarshal customTag")
				continue
			}

			c := wsc.pending[customTag.ID]
			if c == nil {
				wsc.logger.WarnContext(ctx,
					"Reader response with specified customTag.ID does not correspond to any pending request",
					"customTag", customTag.ID)
				continue
			}
			// assign response to call
			c.Resp = response

			select {
			// confirm sending response to Writer
			case c.Done <- true:
				wsc.mutex.Lock()
				// remove call from map as response is already assigned
				delete(wsc.pending, c.ID)
				wsc.mutex.Unlock()
			case <-chanBreaker.C:
				wsc.logger.InfoContext(ctx, "Writer waiting for channel response timeout - protection against goroutine leak")
				return
			// *http.Request context is done after processing request
			// it can be any passed context!
			case <-ctx.Done():
				wsc.logger.InfoContext(ctx, "Context done - reader operation canceled", logging.ErrorAttr(ctx.Err()))
				return
			}
		}
	}
}

// what about Done chan size 1?
type call struct {
	ID   string
	Req  []byte
	Resp []byte
	Done chan bool
}

func (wsc *WSClient) WriteText(ctx context.Context, id string, data []byte) ([]byte, error) {
	// ticker channel has reference to client, so it calculates time immediately after receiving Tick
	// it doesn't start timer when entering this method it starts immediately when client is created
	<-wsc.sendRateLimiter.C

	// TODO
	// log ID?
	wsc.logger.InfoContext(ctx, "Writing message", logging.MsgAttr(string(data)))

	c := &call{
		ID:   id,
		Req:  data,
		Done: make(chan bool),
	}

	// protect against concurrent writes into connection and map
	wsc.mutex.Lock()
	// add request to pending map to match with response
	wsc.pending[c.ID] = c
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	wsc.mutex.Unlock()
	if err != nil {
		return nil, fmt.Errorf("failed to write message by websocket client: %w", err)
	}
	// start timer after writing message
	// TODO
	// maybe instantiate when creating client?
	chanBreaker := time.NewTimer(time.Second * readTimeout)

	select {
	case _, ok := <-c.Done:
		// TODO
		// can also add error to call struct to return error cause from reader
		if !ok {
			return nil, fmt.Errorf("failed to read a response")
		}
		return c.Resp, nil
	case <-chanBreaker.C:
		wsc.logger.WarnContext(ctx, "Writer waiting for channel response timeout - protection against goroutine leak")
		return nil, fmt.Errorf("writer timeout")
	// *http.Request context is done after processing request
	// it can be any passed context!
	case <-ctx.Done():
		err := ctx.Err()
		wsc.logger.InfoContext(ctx, "Context done - writer operation canceled", logging.ErrorAttr(err))
		return nil, fmt.Errorf("writer operation canceled: context done: %w", err)
	}
}
