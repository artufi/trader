package websocket

import (
	"context"
	"fmt"
	"github.com/artufi/trader/config"
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
		dialer:      dialer,
		logger:      logger,
		userClients: make(map[string]map[*WSClient]struct{}),
	}
}

type WSManager struct {
	cfg         config.AppConfig
	dialer      *websocket.Dialer
	logger      *slog.Logger
	userClients map[string]map[*WSClient]struct{}
	sync.RWMutex
}

// DialForNewClient creates a new client
func (wsm *WSManager) DialForNewClient(ctx context.Context, url string, requestHeader http.Header, userID string) (*WSClient, error) {
	conn, resp, err := wsm.dialer.Dial(url, requestHeader)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
			wsm.logger.ErrorContext(ctx, "Failed to dial WebSocket", logging.ErrorAttr(err), logging.URLAttr(url),
				"statusCode", resp.StatusCode)
		} else {
			wsm.logger.ErrorContext(ctx, "Failed to dial WebSocket", logging.ErrorAttr(err), logging.URLAttr(url))
		}
		return nil, fmt.Errorf("failed to establish Websocket connection: %w", err)
	}

	client := &WSClient{
		conn:            conn,
		manager:         wsm,
		logger:          wsm.logger,
		sendRateLimiter: time.NewTicker(200 * time.Millisecond),
		pending:         make(map[string]*call),
		UserID:          userID,
		ReConnCh:        make(chan bool, 1),
	}

	wsm.addClient(ctx, client)

	// read client messages
	go client.ReadMessages(ctx)

	return client, nil
}

func (wsm *WSManager) addClient(ctx context.Context, client *WSClient) {
	wsm.Lock()
	defer wsm.Unlock()

	wsm.logger.InfoContext(ctx, "Adding a new client")
	if clients, ok := wsm.userClients[client.UserID]; ok {
		clients[client] = struct{}{}
	} else {
		wsm.userClients[client.UserID] = make(map[*WSClient]struct{})
		wsm.userClients[client.UserID][client] = struct{}{}
	}
}

func (wsm *WSManager) RemoveClient(ctx context.Context, client *WSClient) {
	wsm.Lock()
	defer wsm.Unlock()

	if clients, ok := wsm.userClients[client.UserID]; ok {
		if _, ok := clients[client]; ok {
			wsm.logger.InfoContext(ctx, "Disconnecting one of user's client", logging.UserIDAttr(client.UserID), client)
			client.CloseConnection()
			client.StopSendRateLimiter()
			delete(clients, client)
		}
	}
}
