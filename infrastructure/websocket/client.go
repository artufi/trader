package websocket

import "github.com/gorilla/websocket"

type WSClient struct {
	conn    *websocket.Conn
	manager *WSManager
	ip      string
	userID  string
}

func (wsc *WSClient) Close() {
	wsc.conn.Close()
}

//func (wsc *WSClient) ReadMessage() error {
//	msgType, bytes, err := wsc.conn.ReadMessage()
//	return err
//}

func (wsc *WSClient) WriteText(data []byte) error {
	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	return err
}
