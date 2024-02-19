package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"log/slog"
	"trader/logging"
)

type WSClient struct {
	conn    *websocket.Conn
	manager *WSManager
	logger  *slog.Logger
	ip      string
	userID  string
}

func (wsc *WSClient) Close() {
	wsc.conn.Close()
}

func (wsc *WSClient) ReadMessages(ctx context.Context) {
	defer func() {
		wsc.manager.RemoveClient(ctx, wsc)
	}()

	for {
		messageType, payload, err := wsc.conn.ReadMessage()
		if err != nil {
			//websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure)
			if err != nil {
				wsc.logger.ErrorContext(ctx, "Failed to read a message", logging.ErrorAttr(err))
			}
			break
		}
		wsc.logger.InfoContext(ctx, "Response message", "messageType", messageType, "payload", string(payload))
	}
}

func (wsc *WSClient) WriteText(data []byte) error {
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	return err
}
