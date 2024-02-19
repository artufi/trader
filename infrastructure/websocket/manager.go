package websocket

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"sync"
	"trader/controller/middleware"
	"trader/logging"
)

func NewWSManager(dialer *websocket.Dialer, logger *slog.Logger) *WSManager {
	if dialer == nil {
		dialer = &websocket.Dialer{}
	}
	return &WSManager{
		dialer:  dialer,
		logger:  logger,
		clients: make(map[*WSClient]bool),
	}
}

type WSManager struct {
	dialer  *websocket.Dialer
	logger  *slog.Logger
	clients map[*WSClient]bool
	sync.RWMutex
}

// Dial creates a new client
func (wsh *WSManager) Dial(ctx context.Context, url string, requestHeader http.Header) (*WSClient, error) {
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
		conn:    conn,
		manager: wsh,
		logger:  wsh.logger,
		userID:  connDetails.UserID,
		ip:      connDetails.IPAddr,
	}
	wsh.addClient(ctx, client)

	return client, nil
}

func (wsh *WSManager) addClient(ctx context.Context, client *WSClient) {
	wsh.Lock()
	defer wsh.Unlock()

	wsh.logger.InfoContext(ctx, "Adding new client")
	wsh.clients[client] = true
}

func (wsh *WSManager) RemoveClient(ctx context.Context, client *WSClient) {
	wsh.Lock()
	defer wsh.Unlock()

	if _, ok := wsh.clients[client]; ok {
		wsh.logger.InfoContext(ctx, "Removing client")
		client.Close()
		delete(wsh.clients, client)
	}
}
