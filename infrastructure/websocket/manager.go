package websocket

import (
	"context"
	"fmt"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/logging"
	"github.com/gorilla/websocket"
	"log/slog"
	"math/rand"
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

// DialForNewClient creates a new client and starts reading messages by this client
func (wsm *WSManager) DialForNewClient(ctx context.Context, url string, requestHeader http.Header, userID string) (*WSClient, error) {
	conn, resp, err := wsm.dialer.Dial(url, requestHeader)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
			wsm.logger.ErrorContext(ctx, "Failed to dial websocket", logging.ErrorAttr(err), logging.URLAttr(url),
				"statusCode", resp.StatusCode)
		} else {
			wsm.logger.ErrorContext(ctx, "Failed to dial websocket", logging.ErrorAttr(err), logging.URLAttr(url))
		}
		return nil, fmt.Errorf("manager failed to establish websocket connection: %w", err)
	}

	client := &WSClient{
		conn:            conn,
		manager:         wsm,
		logger:          wsm.logger,
		sendRateLimiter: time.NewTicker(200 * time.Millisecond),
		pending:         make(map[string]*call),
		UserID:          userID,
		ReConnCh:        make(chan bool),
	}

	wsm.addClient(ctx, client)

	// read client messages
	go client.ReadMessages(ctx)

	return client, nil
}

func (wsm *WSManager) DialForNewStreamClient(ctx context.Context, url string, requestHeader http.Header, userID string) (*WSClientStream, error) {
	conn, resp, err := wsm.dialer.Dial(url, requestHeader)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
			wsm.logger.ErrorContext(ctx, "Manager failed to dial stream websocket", logging.ErrorAttr(err), logging.URLAttr(url),
				"statusCode", resp.StatusCode)
		} else {
			wsm.logger.ErrorContext(ctx, "Manager failed to dial stream websocket", logging.ErrorAttr(err), logging.URLAttr(url))
		}
		return nil, fmt.Errorf("manager failed to establish stream websocket connection: %w", err)
	}

	randomClient, err := wsm.GetUserRandClientWithFreeSSID(userID)
	if err != nil {
		return nil, fmt.Errorf("manager failed to get client for stream connection: %w", err)
	}
	//randomClient.Stream = true
	client := &WSClientStream{
		conn:            conn,
		manager:         wsm,
		logger:          wsm.logger,
		sendRateLimiter: time.NewTicker(200 * time.Millisecond),
		UserID:          userID,
		StreamSessionID: randomClient.StreamSessionID,
		client:          randomClient,
	}

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
			wsm.logger.InfoContext(ctx, "Disconnecting one of user's client")
			client.CloseConnection()
			client.StopSendRateLimiter()
			delete(clients, client)
		}
	}
}

func (wsm *WSManager) GetUserRandomClient(userID string) (*WSClient, error) {
	if clients, ok := wsm.userClients[userID]; ok {
		var clientList []*WSClient
		for client := range clients {
			clientList = append(clientList, client)
		}
		if len(clientList) > 0 {
			return clientList[rand.Intn(len(clientList))], nil
		}
	}
	return nil, fmt.Errorf("manager no clients for user: %v", userID)
}

func (wsm *WSManager) GetUserRandClientWithFreeSSID(userID string) (*WSClient, error) {
	if clients, ok := wsm.userClients[userID]; ok {
		for client := range clients {
			if !client.Stream {
				return client, nil
			}
		}
	}
	return nil, fmt.Errorf("manager no free streaming clients for user: %v", userID)
}
