package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/gorilla/websocket"
	"log/slog"
	"sync"
	"time"
)

const (
	receiverTimeout = 10
	// Time allowed to write a message to the peer.
	writeWait = time.Second * 5
	// Time allowed to read a message from the peer.
	//readWait = time.Second * 60

	// Maximum message size allowed from peer.
	// 100 KB
	maxMessageSize = 102400
)

type WSClient struct {
	ConnID int
	conn   *websocket.Conn

	manager *WSManager
	logger  *slog.Logger

	// User should send requests in 200 ms intervals.
	// This rule can be broken, but if it happens 6 times in a row the connection is dropped.
	sendRateLimiter *time.Ticker

	// WebSocket connection is only allowed to have one concurrent writer,
	// mutex acts also as a locker for pending map during put and delete operations.
	mutex   sync.Mutex
	pending map[string]*call

	UserID          string
	StreamSessionID string
	Stream          bool

	// Buffer size 1 is required to not block client ReadMessages
	ReConnCh chan bool
}

func (wsc *WSClient) CloseConnection() {
	wsc.conn.Close()
}

func (wsc *WSClient) StopSendRateLimiter() {
	wsc.sendRateLimiter.Stop()
}

func (wsc *WSClient) ReadMessages(ctx context.Context) {
	// general reader error that cannot be directly assigned to a specific call (independent of a client write)
	var readerErr error

	defer func() {
		// send signal to reconnect client
		close(wsc.ReConnCh)
		wsc.manager.RemoveClient(ctx, wsc)

		wsc.mutex.Lock()
		for _, c := range wsc.pending {
			if readerErr != nil {
				c.Err = readerErr
			}
			close(c.Done)
			delete(wsc.pending, c.ID)
		}
		wsc.mutex.Unlock()
	}()

	// setup reader
	wsc.conn.SetReadLimit(maxMessageSize)
	//wsc.conn.SetReadDeadline(time.Now().Add(readWait))

	chanBreaker := time.NewTimer(time.Second * receiverTimeout)
	for {
		select {
		// context done can be triggered before any message is read
		case <-ctx.Done():
			readerErr = fmt.Errorf("client reader context done before reading messages: %w", ctx.Err())
			return
		// conn.ReadMessage is blocking so default case here is not the best pick
		// but context can be done before Read, so it makes sense
		default:
			// blocking method
			_, response, err := wsc.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					wsc.logger.ErrorContext(ctx, "Client reader failed to read a message", logging.ErrorAttr(err))
				}
				// For example after Logout: websocket: close 1006 (abnormal closure): unexpected EOF
				// Logout is normally processed, response (ReadMessage) is read correctly, and after
				// processing Logout conn (connection) is closed with the close 1006 message.
				// This situation also occur after too many Writes to XTB by client.
				readerErr = fmt.Errorf("client reader failure: %w", err)
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
			chanBreaker.Reset(time.Second * receiverTimeout)

			// read customTag from response to link to request
			customTag := struct {
				ID string `json:"customTag"`
			}{}
			err = json.Unmarshal(response, &customTag)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Client reader unable to unmarshal customTag", logging.ErrorAttr(err))
				continue
			}

			c := wsc.pending[customTag.ID]
			if c == nil {
				wsc.logger.ErrorContext(ctx,
					"Client reader response with specified customTag.ID does not correspond to any pending request",
					"customTag", customTag.ID)
				continue
			}
			// assign response to call
			c.Resp = response

			// c.Done is a buffered channel of 1, other cases than the c.Done <- true are even impossible to occur
			// if no one is waiting to receive from the c.Done, then Go runtime or defer will clear resources
			select {
			// confirm sending response to Writer
			case c.Done <- true:
				wsc.mutex.Lock()
				// remove call from map as response is already assigned
				delete(wsc.pending, c.ID)
				wsc.mutex.Unlock()

			// only possible if a client writer stops waiting for the response - client will receive error cause
			// or a client reader timeout is so short that chanBreaker triggers before sending the response - client won't receive error cause
			case <-chanBreaker.C:
				wsc.logger.WarnContext(ctx, "Client reader waiting for channel response timeout - protection against client reader block")
				c.Err = fmt.Errorf("client reader timeout while waiting for accepting response by writer: " +
					"protection against client reader block")
				// close and remove from map to prevent panic when closing channel in defer
				wsc.closeAndDelete(c)

			// context done can be triggered while sending response
			case <-ctx.Done():
				wsc.logger.WarnContext(ctx, "Client reader operation canceled - context done", logging.ErrorAttr(ctx.Err()))
				c.Err = fmt.Errorf("client reader operation canceled - context done: %w", ctx.Err())
				// close and remove from map to prevent panic when closing channel in defer
				wsc.closeAndDelete(c)
			}
		}
	}
}

func (wsc *WSClient) closeAndDelete(c *call) {
	close(c.Done)
	wsc.mutex.Lock()
	delete(wsc.pending, c.ID)
	wsc.mutex.Unlock()
}

// Done is a buffered chan with size 1, it prevents blocking of the client reader
type call struct {
	ID   string
	Req  []byte
	Resp []byte
	Done chan bool
	Err  error
}

func (wsc *WSClient) WriteText(ctx context.Context, messageID string, data []byte) ([]byte, error) {
	select {
	// *http.Request context is done after processing request
	// it can be any passed context
	// do not process message if context is done before writing message
	case <-ctx.Done():
		return nil, fmt.Errorf("client writer context done before writing message: %w", ctx.Err())

	// ticker channel has reference to client, so it calculates time immediately after receiving Tick
	// it doesn't start timer when entering this method it starts immediately when client is created
	case <-wsc.sendRateLimiter.C:
		wsc.logger.InfoContext(ctx, "Writing message", logging.MsgAttr(string(data)))

		c := &call{
			ID:  messageID,
			Req: data,
			// buffered chan with size 1, it prevents blocking of the client reader
			Done: make(chan bool, 1),
		}

		// protect against concurrent writes into connection and map
		wsc.mutex.Lock()
		// add request to pending map to match with response
		wsc.pending[c.ID] = c
		wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := wsc.conn.WriteMessage(websocket.TextMessage, data)
		wsc.mutex.Unlock()
		if err != nil {
			return nil, fmt.Errorf("client writer failed to write message through websocket: %w", err)
		}

		// TODO
		// maybe instantiate when creating client?
		// maybe context.WithTimeout instead of chanBreaker?
		// start timer after writing message
		chanBreaker := time.NewTimer(time.Second * receiverTimeout)

		select {
		case _, ok := <-c.Done:
			if !ok {
				return nil, fmt.Errorf("client writer failed to read a response: %w", c.Err)
			}
			return c.Resp, nil
		case <-chanBreaker.C:
			return nil, fmt.Errorf("client writer timeout while waiting for channel response: protection against goroutine leak")

		// context done can be triggered while waiting for a response
		case <-ctx.Done():
			return nil, fmt.Errorf("client writer context done while waiting for a response: %w", ctx.Err())
		}
	}
}

// Ping is better to be invoked outside of manager to pass context with more information as manager
// does not have all of them when Dialing for a new client
func (wsc *WSClient) Ping(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(time.Second * interval)
	for {
		select {
		case <-ticker.C:
			wsc.logger.InfoContext(ctx, "Ping")
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := wsc.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Client failed to write Ping message", logging.ErrorAttr(err))
				return
			}
		case <-ctx.Done():
			err := ctx.Err()
			wsc.logger.ErrorContext(ctx, "Client Ping operation canceled - context done", logging.ErrorAttr(err))
			return
		}
	}
}

type WSClientStream struct {
	ConnID int
	conn   *websocket.Conn

	client  *WSClient
	manager *WSManager
	logger  *slog.Logger

	// User should send requests in 200 ms intervals.
	// This rule can be broken, but if it happens 6 times in a row the connection is dropped.
	sendRateLimiter *time.Ticker

	UserID          string
	StreamSessionID string
}

func (wsc *WSClientStream) FreeClient() {
	if wsc.client != nil {
		wsc.client.Stream = false
	}
}

func (wsc *WSClientStream) CloseConnection() {
	wsc.conn.Close()
}

func (wsc *WSClientStream) StopSendRateLimiter() {
	wsc.sendRateLimiter.Stop()
}

func (wsc *WSClientStream) ReadMessages(ctx context.Context) ([]byte, error) {
	// setup reader
	//wsc.conn.SetReadLimit(maxMessageSize)
	//wsc.conn.SetReadDeadline(time.Now().Add(readWait))

	for {
		select {
		// context done can be triggered before any message is read
		case <-ctx.Done():
			return nil, fmt.Errorf("client stream reader context done before reading messages: %w", ctx.Err())
		// conn.ReadMessage is blocking so default case here is not the best pick
		// but context can be done before Read, so it makes sense
		default:
			// blocking method
			// to better fit with ctx.Done use goroutine
			_, response, err := wsc.conn.ReadMessage()
			if err != nil {
				return nil, fmt.Errorf("client stream reader failure: %w", err)
			}

			select {
			// context is done, no one want response
			case <-ctx.Done():
				return nil, fmt.Errorf("client stream reader operation canceled - context done: %w", ctx.Err())
			default:
				return response, nil
			}
		}
	}
}

func (wsc *WSClientStream) WriteText(ctx context.Context, messageID string, data []byte) ([]byte, error) {
	select {
	// do not process message if context is done before writing message
	case <-ctx.Done():
		return nil, fmt.Errorf("client stream writer context done before writing message: %w", ctx.Err())

	// ticker channel has reference to client, so it calculates time immediately after receiving Tick
	// it doesn't start timer when entering this method it starts immediately when client is created
	case <-wsc.sendRateLimiter.C:
		wsc.logger.InfoContext(ctx, "Writing stream message", logging.MsgAttr(string(data)))

		wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := wsc.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			return nil, fmt.Errorf("client stream writer failed to write message through websocket: %w", err)
		}
		return nil, nil
	}
}

func (wsc *WSClientStream) Ping(ctx context.Context, interval time.Duration, ssid string) {
	ticker := time.NewTicker(time.Second * interval)
	for {
		select {
		case <-ticker.C:
			wsc.logger.InfoContext(ctx, "Ping stream")
			pingStreamJSON, err := jsonform.PingStream(ssid)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Failed to serialize pingStream", logging.ErrorAttr(err))
				return
			}
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = wsc.conn.WriteMessage(websocket.TextMessage, pingStreamJSON)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Client stream failed to write Ping message", logging.ErrorAttr(err))
				return
			}
		case <-ctx.Done():
			err := ctx.Err()
			wsc.logger.ErrorContext(ctx, "Client stream Ping operation canceled - context done", logging.ErrorAttr(err))
			return
		}
	}
}
