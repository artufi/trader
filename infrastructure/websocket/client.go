package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"log/slog"
	"time"
	"trader/logging"
)

const readTimeout = 10

type WSClient struct {
	conn    *websocket.Conn
	manager *WSManager
	logger  *slog.Logger

	// User should send requests in 200 ms intervals.
	// This rule can be broken, but if it happens 6 times in a row the connection is dropped.
	sendRateLimiter *time.Ticker

	ip     string
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

// TODO
// what about respCh to have size 1?
func (wsc *WSClient) ReadMessages(ctx context.Context, respCh chan<- []byte) {
	defer func() {
		wsc.manager.RemoveClient(ctx, wsc)
	}()

	chanBreaker := time.NewTimer(time.Second * readTimeout)
	for {
		select {
		// *http.Request context is done after processing request
		// it can be any passed context!
		case <-ctx.Done():
			wsc.logger.InfoContext(ctx, "Context done")
			// context canceled somewhere so better close channel in case some goroutine waits for a message
			close(respCh)
			return
		// TODO
		// conn.ReadMessage is blocking so default case here is not the best pick
		default:
			// blocking method
			_, response, err := wsc.conn.ReadMessage()
			if err != nil {
				//websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure)
				wsc.logger.ErrorContext(ctx, "Failed to read a message", logging.ErrorAttr(err))
				close(respCh)
				return
			}
			// TODO
			// don't need to Stop timer, as it invokes close after specified time
			// of course it is possible to Stop as Timer won't be instantly garbage collected or never?
			chanBreaker.Reset(time.Second * readTimeout)

			select {
			// just send response
			case respCh <- response:
			case <-chanBreaker.C:
				wsc.logger.InfoContext(ctx, "Response channel timeout - protection against goroutine leak")
				// no need to close respCh as no one waits to receive a message
				close(respCh)
				return
			// *http.Request context is done after processing request
			// it can be any passed context!
			case <-ctx.Done():
				wsc.logger.InfoContext(ctx, "Context done")
				// context canceled somewhere so better close channel in case some goroutine waits for a message
				close(respCh)
				return
			}
		}
	}
}

func (wsc *WSClient) WriteText(ctx context.Context, data []byte) error {
	// channel has reference to client, so it calculates time immediately after receiving Tick
	// it doesn't start timer when entering this method
	<-wsc.sendRateLimiter.C

	wsc.logger.InfoContext(ctx, "Writing message", logging.MsgAttr(string(data)))
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	return err
}
