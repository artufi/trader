package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"sync"
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
func (wsh *WSManager) Dial(url string, requestHeader http.Header, reqIP string, userID string) (*WSClient, error) {
	conn, resp, err := wsh.dialer.Dial(url, requestHeader)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
			wsh.logger.Error("Failed to dial WebSocket", logging.ErrorAttr(err), logging.URLAttr(url),
				"statusCode", resp.StatusCode)
		} else {
			wsh.logger.Error("Failed to dial WebSocket", logging.ErrorAttr(err), logging.URLAttr(url))
		}
		return nil, fmt.Errorf("failed to establish Websocket connection: %w", err)
	}

	client := &WSClient{
		conn:    conn,
		manager: wsh,
		ip:      reqIP,
		userID:  userID,
	}
	wsh.addClient(client)

	return client, nil
}

func (wsh *WSManager) addClient(client *WSClient) {
	wsh.Lock()
	defer wsh.Unlock()

	wsh.logger.Info("Adding new client", logging.UserIDAttr(client.userID), logging.ClientIPAttr(client.ip))
	wsh.clients[client] = true
}

func (wsh *WSManager) RemoveClient(client *WSClient) {
	wsh.Lock()
	defer wsh.Unlock()

	if _, ok := wsh.clients[client]; ok {
		wsh.logger.Info("Removing client", logging.UserIDAttr(client.userID), logging.ClientIPAttr(client.ip))
		client.Close()
		delete(wsh.clients, client)
	}
}
