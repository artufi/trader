package websocket

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestWSClient_WriteText_ContextDoneBeforeWritingMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer srv.Close()

	conn := dialWS(t, srv)
	defer conn.Close()
	wsClient := newTestWSClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := wsClient.WriteText(ctx, "msg1", []byte(`{"test":"data"}`))
	if err == nil {
		t.Errorf("expected client writer context done before writing message, got: nil")
	}
}

func TestWSClient_WriteText_WriteMessageFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer srv.Close()

	conn := dialWS(t, srv)
	wsClient := newTestWSClient(conn)
	conn.Close()

	_, err := wsClient.WriteText(context.Background(), "testMsgID", []byte(`{"customTag":"msg"}`))
	if err == nil {
		t.Errorf("expected error due to conn.WriteMessage failure, got: nil")
	}
}

func TestWSClient_WriteText_ContextTimeoutWhileWaitingForResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer srv.Close()

	conn := dialWS(t, srv)
	defer conn.Close()
	wsClient := newTestWSClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := wsClient.WriteText(ctx, "testMsgID", []byte(`{"test":"data"}`))
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context deadline exceeded, got: %v", err)
	}
}

func TestWSClient_WriteText_ContextCancelledAfterWrite(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer srv.Close()

	conn := dialWS(t, srv)
	defer conn.Close()
	wsClient := newTestWSClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			_, ok := wsClient.pending["testMessageID"]
			if ok {
				cancel()
				return
			}
		}
	}()

	_, err := wsClient.WriteText(ctx, "testMessageID", []byte(`{"x":1}`))
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("expected context canceled, got: %v", err)
	}
}

func TestWSClient_WriteText_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer srv.Close()

	conn := dialWS(t, srv)
	defer conn.Close()
	wsClient := newTestWSClient(conn)

	expectedResp := "testResponse"
	done := make(chan struct{})
	var resp []byte
	var err error
	messageID := "testMsgID"
	go func() {
		resp, err = wsClient.WriteText(context.Background(), messageID, []byte(expectedResp))
		close(done)
	}()

	for {
		c, ok := wsClient.pending[messageID]
		if ok {
			c.Resp = []byte(expectedResp)
			c.Done <- true
			break
		}
	}
	<-done

	if err != nil {
		t.Fatal(err)
	}
	if expectedResp != string(resp) {
		t.Errorf("expected: %s, got: %s", expectedResp, string(resp))
	}
}
