package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/gorilla/websocket"
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
	// Indicate that logged-in client is also used in stream mode.
	Stream bool

	// Buffer size 1 is required to not block client ReadMessages.
	ReConnCh chan bool
}

func (wsc *WSClient) CloseConnection() {
	wsc.conn.Close()
}

func (wsc *WSClient) StopSendRateLimiter() {
	wsc.sendRateLimiter.Stop()
}

func (wsc *WSClient) ReadMessages(ctx context.Context) {
	// General reader error that cannot be directly assigned to a specific call (independent of a client write).
	var readerErr error

	defer func() {
		// Send signal to reconnect client.
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

	// Setup reader limits.
	wsc.conn.SetReadLimit(maxMessageSize)
	//wsc.conn.SetReadDeadline(time.Now().Add(readWait))

	chanBreaker := time.NewTimer(time.Second * receiverTimeout)
	for {
		select {
		// Context done can be triggered before any message is read.
		case <-ctx.Done():
			readerErr = fmt.Errorf("client reader context done before reading messages: %w", ctx.Err())
			return
		default:
			// TODO:
			// Since ReadMessage is blocking, to better respond to ctx.Done cancellation,
			// run ReadMessage in a goroutine, e.g.,
			// go func() { _, message, err := wsc.conn.ReadMessage(); msgCh <- message }()
			// This allows non-blocking reads and sends the message to a channel when data arrives.
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

			// There is a minimal chance that the Stop method may not detect that the timer
			// has expired or was stopped because it was called just before the timer's expiration.
			// Consequently, Reset may return false, and the timer will not be cleared.
			// However, this is super rare.
			// ---
			// This step prevents the immediate reading from the timer's channel and
			// allows for the timer to be reset for the next Read operation.
			chanBreaker.Reset(time.Second * receiverTimeout)

			// Read customTag from response to link to request.
			customTag := struct {
				ID string `json:"customTag"`
			}{}
			err = json.Unmarshal(response, &customTag)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Client reader unable to unmarshal customTag", logging.ErrorAttr(err))
				continue
			}

			wsc.mutex.Lock()
			c := wsc.pending[customTag.ID]
			wsc.mutex.Unlock()
			if c == nil {
				wsc.logger.ErrorContext(ctx,
					"Client reader response with specified customTag.ID does not correspond to any pending request",
					"customTag", customTag.ID)
				continue
			}
			// Assign response to pending call.
			c.Resp = response

			// c.Done is a buffered channel of 1, other cases than the c.Done <- true are even impossible to occur
			// if no one is waiting to receive from the c.Done, then Go runtime or defer will clear resources.
			select {
			// Confirm sending response to Writer.
			case c.Done <- true:
				wsc.mutex.Lock()
				// Remove call from map as response is already assigned.
				delete(wsc.pending, c.ID)
				wsc.mutex.Unlock()
			// Only possible if a client writer stops waiting for the response - client won't receive error cause as no one is waiting for it
			// or a client reader timeout is so short that chanBreaker triggers before sending the response - client will receive error cause
			// or somebody changes a buffer size of the call.Done channel to 0 (processing time increase) - client will receive error cause.
			case <-chanBreaker.C:
				wsc.logger.WarnContext(ctx, "Client reader waiting for channel response timeout - protection against client reader block")
				c.Err = fmt.Errorf("client reader timeout while waiting for accepting response by writer: " +
					"protection against client reader block")
				// Close and remove from map to prevent panic when closing channel in defer.
				wsc.closeAndDelete(c)
			case <-ctx.Done():
				wsc.logger.WarnContext(ctx, "Client reader operation canceled - context done", logging.ErrorAttr(ctx.Err()))
				c.Err = fmt.Errorf("client reader operation canceled - context done: %w", ctx.Err())
				// Close and remove from map to prevent panic when closing channel in defer.
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

// Done is a buffered chan with size 1, it prevents blocking of the client reader.
type call struct {
	ID   string
	Req  []byte
	Resp []byte
	Done chan bool
	Err  error
}

func (wsc *WSClient) WriteText(ctx context.Context, messageID string, data []byte) ([]byte, error) {
	select {
	// *http.Request context is done after processing request.
	// Tt can be any passed context.
	// Do not process message if context is done before writing message.
	case <-ctx.Done():
		return nil, fmt.Errorf("client writer context done before writing message: %w", ctx.Err())

	// Ticker channel has reference to client, so it calculates time immediately after receiving Tick.
	// It doesn't start timer when entering this method it starts immediately when client is created.
	case <-wsc.sendRateLimiter.C:
		wsc.logger.InfoContext(ctx, "Writing message", logging.MsgAttr(string(data)))

		c := &call{
			ID:  messageID,
			Req: data,
			// Buffered chan with size 1, it prevents blocking of the client reader.
			Done: make(chan bool, 1),
		}

		// Protect against concurrent writes into connection and map.
		wsc.mutex.Lock()
		// Add request to pending map to match with response.
		wsc.pending[c.ID] = c
		wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := wsc.conn.WriteMessage(websocket.TextMessage, data)
		wsc.mutex.Unlock()
		if err != nil {
			return nil, fmt.Errorf("client writer failed to write message through websocket: %w", err)
		}

		// Start timer after writing message.
		ctx, cancel := context.WithTimeout(ctx, time.Second*receiverTimeout)
		defer cancel()

		select {
		case _, ok := <-c.Done:
			if !ok {
				return nil, fmt.Errorf("client writer failed to read a response: %w", c.Err)
			}
			return c.Resp, nil
		case <-ctx.Done():
			err = ctx.Err()
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("client writer timeout while waiting for a response: %w", err)
			}
			return nil, fmt.Errorf("client writer context done while waiting for a response: %w", err)
		}
	}
}

// Ping is better to be invoked outside of manager to pass context with more information as manager
// does not have all of them when Dialing for a new client.
func (wsc *WSClient) Ping(ctx context.Context, interval time.Duration) {
	if interval < 1 {
		wsc.logger.WarnContext(ctx, "Too low ping interval value", "value", interval)
		return
	}
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
			wsc.logger.InfoContext(ctx, "Client Ping operation canceled - context done")
			return
		}
	}
}

// WSClientStream highly depends on logged in WSClient.
type WSClientStream struct {
	ConnID int
	conn   *websocket.Conn

	// WSClientStream must be bind to logged in *WSClient with free SSID to release
	// reserved *WSClient in case of dropped streaming connection by WSClientStream
	// due to fault and do not block particular *WSClient streamSessionId to be
	// used by other WSClientStream.
	client  *WSClient
	manager *WSManager
	logger  *slog.Logger

	// User should send requests in 200 ms intervals.
	// This rule can be broken, but if it happens 6 times in a row the connection is dropped.
	sendRateLimiter *time.Ticker

	UserID          string
	StreamSessionID string

	// Response from ReadMessages send by WSClientStream.
	ReaderRespCh chan ReaderResp
}

type ReaderResp struct {
	Data []byte
	Err  error
}

// ReleaseClientStreamResources releases *WSClient stream capabilities (turn off stream mode)
// allowing other WSClientStream reuse particular *WSClient streamSessionId if available.
func (wsc *WSClientStream) ReleaseClientStreamResources() {
	// Client can be nil if something happened with logged in *WSClient and GC.
	// Cleans resources and removes it from memory.
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

func (wsc *WSClientStream) ReadMessages(ctx context.Context) {
	var readerErr error
	// Setup reader.
	//wsc.conn.SetReadLimit(maxMessageSize)
	//wsc.conn.SetReadDeadline(time.Now().Add(readWait))

	defer func() {
		wsc.ReaderRespCh <- ReaderResp{
			Err: readerErr,
		}
		wsc.ReleaseClientStreamResources()
	}()

	chanBreaker := time.NewTimer(time.Second * receiverTimeout)
	for {
		select {
		// Context done can be triggered before any message is read.
		case <-ctx.Done():
			readerErr = fmt.Errorf("client-stream reader context done before reading messages: %w", ctx.Err())
			return
		default:
			// TODO:
			// Since ReadMessage is blocking, to better respond to ctx.Done cancellation,
			// run ReadMessage in a goroutine, e.g.,
			// go func() { _, message, err := wsc.conn.ReadMessage(); msgCh <- message }()
			// This allows non-blocking reads and sends the message to a channel when data arrives.
			_, response, err := wsc.conn.ReadMessage()
			if err != nil {
				readerErr = fmt.Errorf("client-stream reader failure: %w", err)
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

			// There is a minimal chance that the Stop method may not detect that the timer
			// has expired or was stopped because it was called just before the timer's expiration.
			// Consequently, Reset may return false, and the timer will not be cleared.
			// However, this is super rare.
			// ---
			// This step prevents the immediate reading from the timer's channel and
			// allows for the timer to be reset for the next Read operation.
			chanBreaker.Reset(time.Second * receiverTimeout)

			select {
			case wsc.ReaderRespCh <- ReaderResp{Data: response}:
			case <-chanBreaker.C:
				wsc.logger.WarnContext(ctx, "Client-stream reader waiting for channel response timeout - protection against client reader block")
				readerErr = fmt.Errorf("client-stream reader timeout while waiting for accepting response by writer: " +
					"protection against client reader block")
			// Context is done, no one wants response.
			case <-ctx.Done():
				readerErr = fmt.Errorf("client-stream reader operation canceled - context done: %w", ctx.Err())
				return
			}
		}
	}
}

func (wsc *WSClientStream) WriteText(ctx context.Context, messageID string, data []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("client-stream writer context done before writing message: %w", ctx.Err())
	// Ticker channel has reference to client, so it calculates time immediately after receiving Tick
	// it doesn't start timer when entering this method it starts immediately when client is created.
	case <-wsc.sendRateLimiter.C:
		wsc.logger.InfoContext(ctx, "Writing stream message", logging.MsgAttr(string(data)))

		wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := wsc.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			return nil, fmt.Errorf("client-stream writer failed to write message through websocket: %w", err)
		}
		return nil, nil
	}
}

func (wsc *WSClientStream) Ping(ctx context.Context, interval time.Duration) {
	if interval < 1 {
		wsc.logger.WarnContext(ctx, "Client-stream too low ping interval value", logging.IntervalAttr(interval))
		return
	}
	ticker := time.NewTicker(time.Second * interval)
	for {
		select {
		case <-ticker.C:
			wsc.logger.InfoContext(ctx, "Client-stream Ping")
			pingStreamJSON, err := jsonform.PingStream(wsc.StreamSessionID)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Client-stream failed to serialize pingStream", logging.ErrorAttr(err))
				return
			}
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = wsc.conn.WriteMessage(websocket.TextMessage, pingStreamJSON)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Client-stream failed to write Ping message", logging.ErrorAttr(err))
				return
			}
		case <-ctx.Done():
			err := ctx.Err()
			wsc.logger.ErrorContext(ctx, "Client-stream Ping operation canceled - context done", logging.ErrorAttr(err))
			return
		}
	}
}
