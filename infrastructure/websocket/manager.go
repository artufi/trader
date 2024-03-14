package websocket

import (
	"context"
	"errors"
	"fmt"
	"github.com/artufi/trader/controller/middleware"
	"github.com/artufi/trader/logging"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

func NewWSManager(dialer *websocket.Dialer, logger *slog.Logger) *WSManager {
	if dialer == nil {
		dialer = &websocket.Dialer{}
	}
	return &WSManager{
		dialer:  dialer,
		logger:  logger,
		clients: make(map[*WSClient]struct{}),
	}
}

type WSManager struct {
	dialer  *websocket.Dialer
	logger  *slog.Logger
	clients map[*WSClient]struct{}
	sync.RWMutex
}

// DialForNewClient creates a new client
func (wsh *WSManager) DialForNewClient(ctx context.Context, url string, requestHeader http.Header) (*WSClient, error) {
	conn, resp, err := wsh.dialer.Dial(url, requestHeader)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
			wsh.logger.ErrorContext(ctx, "Failed to dial WebSocket", logging.ErrorAttr(err), logging.URLAttr(url), "statusCode", resp.StatusCode)
		} else {
			wsh.logger.ErrorContext(ctx, "Failed to dial WebSocket", logging.ErrorAttr(err), logging.URLAttr(url))
		}
		return nil, fmt.Errorf("failed to establish Websocket connection: %w", err)
	}

	connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails)
	if !ok {
		return nil, errors.New("lack of connection details")
	}

	client := &WSClient{
		conn:            conn,
		manager:         wsh,
		logger:          wsh.logger,
		sendRateLimiter: time.NewTicker(200 * time.Millisecond),
		pending:         make(map[string]*call),
		userID:          connDetails.UserID,
		ip:              connDetails.IPAddr,
	}

	wsh.addClient(ctx, client)

	// read client messages
	go client.ReadMessages(ctx)

	return client, nil
}

func (wsh *WSManager) addClient(ctx context.Context, client *WSClient) {
	wsh.Lock()
	defer wsh.Unlock()

	wsh.logger.InfoContext(ctx, "Adding new client")
	wsh.clients[client] = struct{}{}
}

func (wsh *WSManager) RemoveClient(ctx context.Context, client *WSClient) {
	wsh.Lock()
	defer wsh.Unlock()

	if _, ok := wsh.clients[client]; ok {
		wsh.logger.InfoContext(ctx, "Disconnecting client")
		client.CloseConnection()
		client.StopSendRateLimiter()
		delete(wsh.clients, client)
	}
}
