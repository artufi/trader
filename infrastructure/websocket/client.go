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

func (wsc *WSClient) ReadMessages(ctx context.Context, respCh chan<- []byte) {
	defer func() {
		wsc.manager.RemoveClient(ctx, wsc)
	}()

	for {
		select {
		case <-ctx.Done():
			close(respCh)
			return
		default:
			_, response, err := wsc.conn.ReadMessage()
			if err != nil {
				//websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure)
				wsc.logger.ErrorContext(ctx, "Failed to read a message", logging.ErrorAttr(err))
				close(respCh)
				return
			}
			if len(response) > 0 {
				respCh <- response
			}
		}
	}
}

func (wsc *WSClient) WriteText(data []byte) error {
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	return err
}
