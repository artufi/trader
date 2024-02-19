package websocket

import (
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

func (wsc *WSClient) ReadMessages() {
	defer func() {
		wsc.manager.RemoveClient(wsc)
	}()

	for {
		messageType, payload, err := wsc.conn.ReadMessage()
		if err != nil {
			//websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure)
			if err != nil {
				wsc.logger.Error("Failed to read a message.", logging.ErrorAttr(err))
			}
			break
		}
		wsc.logger.Info("Response message.", "messageType", messageType, "payload", string(payload))
	}
}

func (wsc *WSClient) WriteText(data []byte) error {
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	return err
}
