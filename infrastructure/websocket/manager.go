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

func NewWSManager(cfg config.AppConfig, dialer *websocket.Dialer, logger *slog.Logger) *WSManager {
	if dialer == nil {
		dialer = &websocket.Dialer{}
	}
	return &WSManager{
		cfg:         cfg,
		dialer:      dialer,
		logger:      logger,
		userClients: make(map[string]map[*WSClient]struct{}),
	}
}

type WSManager struct {
	cfg    config.AppConfig
	dialer *websocket.Dialer
	logger *slog.Logger

	// Connections per user.
	userClients map[string]map[*WSClient]struct{}

	// To prevent concurrent goroutines from accessing the same logged-in *WSClient connection,
	// we aim to avoid a map race condition where the CPU may not update memory within a short
	// time gap – meaning that one goroutine could alter a resource, while another might not see the updated state.
	// Using RWMutex with maps is recommended, as concurrent writes and reads are not permissible.

	// Prevent concurrent goroutines access to the same logged in *WSClient connection,
	// it means to prevent map race condition (CPU may not update memory during a short
	// gap - if one goroutine changes a resource another may not see the updated state).
	// RWMutex should be used with map, as concurrent write and read is not ok.
	userClientsMutex sync.RWMutex
}

// DialForNewClient
// creates a new client,
// reads messages.
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

	// Read client messages.
	go client.ReadMessages(ctx)

	return client, nil
}

// DialForNewStreamClient
// creates a new stream client,
// reads messages,
// sends ping messages.
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

	randomFreeClient, err := wsm.GetUserRandClientWithFreeSSID(userID)
	if err != nil {
		return nil, fmt.Errorf("manager failed to get client for stream connection: %w", err)
	}
	ssid := randomFreeClient.StreamSessionID

	client := &WSClientStream{
		conn:            conn,
		manager:         wsm,
		logger:          wsm.logger,
		sendRateLimiter: time.NewTicker(200 * time.Millisecond),
		UserID:          userID,
		StreamSessionID: ssid,
		client:          randomFreeClient,
		// To avoid potential latency increases for the receiver unable to
		// handle the speed and sender unable to read messages from proxy server
		// due to heavy blocking during channel transmission
		// the stream reader response channel cannot be buffered,
		// as it will unavoidably impact one or both sides of the channel message exchange.
		ReaderRespCh: make(chan ReaderResp),
	}

	ctx = logging.AppendAttrsCtx(ctx, logging.StreamID(ssid))
	go client.ReadMessages(ctx)
	// SSID is known at this point, so it is better to ping immediately after creating client.
	// This is different situation than normal WSClient where connections are managed by ConnManager
	// and SSID is not known before log into XTB.
	go client.Ping(ctx, time.Duration(wsm.cfg.Client.Stream.Ping.IntervalSec))

	return client, nil
}

func (wsm *WSManager) addClient(ctx context.Context, client *WSClient) {
	wsm.userClientsMutex.Lock()
	defer wsm.userClientsMutex.Unlock()

	wsm.logger.InfoContext(ctx, "Adding a new client")
	if clients, ok := wsm.userClients[client.UserID]; ok {
		clients[client] = struct{}{}
	} else {
		wsm.userClients[client.UserID] = make(map[*WSClient]struct{})
		wsm.userClients[client.UserID][client] = struct{}{}
	}
}

func (wsm *WSManager) RemoveClient(ctx context.Context, client *WSClient) {
	wsm.userClientsMutex.Lock()
	defer wsm.userClientsMutex.Unlock()

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
	// RWMutex to block access during concurrent read and write.
	// RWMutex (RLock) does not block when there is no lock on write (add, delete).
	wsm.userClientsMutex.RLock()
	defer wsm.userClientsMutex.RUnlock()

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
	wsm.userClientsMutex.Lock()
	defer wsm.userClientsMutex.Unlock()

	if clients, ok := wsm.userClients[userID]; ok {
		for client := range clients {
			if !client.Stream && len(client.StreamSessionID) > 0 {
				client.Stream = true
				return client, nil
			}
		}
	}
	return nil, fmt.Errorf("manager no free streaming clients for user: %v", userID)
}
