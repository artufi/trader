package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func dialWS(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	return conn
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func newTestWSClient(conn *websocket.Conn) *WSClient {
	return &WSClient{
		conn:            conn,
		manager:         &WSManager{},
		logger:          newDiscardLogger(),
		sendRateLimiter: time.NewTicker(200 * time.Millisecond),
		pending:         make(map[string]*call),
		UserID:          "testUserID",
		ReConnCh:        make(chan bool),
	}
}

func newServerClient(t *testing.T, handler http.HandlerFunc) *WSClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	conn := dialWS(t, srv)
	t.Cleanup(func() { conn.Close() })

	return newTestWSClient(conn)
}

func TestWSClient_WriteText(t *testing.T) {
	t.Run("context done before writing", func(t *testing.T) {
		wsClient := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Error(err)
				return
			}
			conn.Close()
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := wsClient.WriteText(ctx, "msg1", []byte(`{"test":"data"}`))
		if err == nil {
			t.Errorf("expected client writer context done before writing message, got: nil")
		}
	})

	t.Run("connection write fails", func(t *testing.T) {
		wsClient := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Error(err)
				return
			}
			conn.Close()
		})
		wsClient.CloseConnection()

		_, err := wsClient.WriteText(context.Background(), "testMsgID", []byte(`{"customTag":"msg"}`))
		if err == nil {
			t.Errorf("expected error due to conn.WriteMessage failure, got: nil")
		}
	})

	t.Run("response times out", func(t *testing.T) {
		wsClient := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Error(err)
				return
			}
			conn.Close()
		})

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := wsClient.WriteText(ctx, "testMsgID", []byte(`{"test":"data"}`))
		if err == nil || !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected context deadline exceeded, got: %v", err)
		}
	})

	t.Run("context cancelled after write", func(t *testing.T) {
		received := make(chan struct{})

		wsClient := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()

			if _, _, err := conn.ReadMessage(); err != nil {
				t.Error(err)
				return
			}
			close(received)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-received
			cancel()
		}()

		_, err := wsClient.WriteText(ctx, "testMessageID", []byte(`{"x":1}`))
		if err == nil || !errors.Is(err, context.Canceled) {
			t.Errorf("expected context canceled, got: %v", err)
		}
	})

	t.Run("server responds with matching message id", func(t *testing.T) {
		errCh := make(chan error, 1)

		wsClient := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
			defer close(errCh)

			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()

			_, msg, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			var message struct {
				ID string `json:"customTag"`
			}
			if err = json.Unmarshal(msg, &message); err != nil {
				errCh <- err
				return
			}
			resp, err := json.Marshal(map[string]any{"customTag": message.ID, "status": true})
			if err != nil {
				errCh <- err
				return
			}
			if err = conn.WriteMessage(websocket.TextMessage, resp); err != nil {
				errCh <- err
			}
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go wsClient.ReadMessages(ctx)

		expectedResp := `{"customTag":"testId","status":true}`
		resp, err := wsClient.WriteText(context.Background(), "testId", []byte(expectedResp))
		if err != nil {
			t.Fatalf("failed to write: %v", err)
		}
		if expectedResp != string(resp) {
			t.Errorf("expected: %s, got: %s", expectedResp, string(resp))
		}
		if srvErr := <-errCh; srvErr != nil {
			t.Fatalf("server error: %v", srvErr)
		}
	})

	t.Run("multiplexes concurrent requests by message id", func(t *testing.T) {
		wsClient := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()

			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				var message struct {
					ID string `json:"customTag"`
				}
				if err := json.Unmarshal(msg, &message); err != nil {
					t.Error(err)
					return
				}
				resp, _ := json.Marshal(map[string]any{"customTag": message.ID})
				if err := conn.WriteMessage(websocket.TextMessage, resp); err != nil {
					t.Error(err)
					return
				}
			}
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go wsClient.ReadMessages(ctx)

		const n = 10
		var wg sync.WaitGroup
		resps := make([][]byte, n)
		errs := make([]error, n)
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				id := fmt.Sprintf("id-%d", i)
				resps[i], errs[i] = wsClient.WriteText(context.Background(), id, []byte(fmt.Sprintf(`{"customTag":%q}`, id)))
			}(i)
		}
		wg.Wait()

		for i := 0; i < n; i++ {
			if errs[i] != nil {
				t.Errorf("request: %d failed: %v", i, errs[i])
				continue
			}
			expected := fmt.Sprintf(`"customTag":"id-%d"`, i)
			if !strings.Contains(string(resps[i]), expected) {
				t.Errorf("request: %d got wrong response: %s", i, resps[i])
			}
		}
	})
}
