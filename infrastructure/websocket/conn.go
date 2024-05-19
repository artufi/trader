package websocket

import (
	"context"
	"fmt"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/xtb/processor"
	"log/slog"
	"time"
)

const serviceName = "ConnManager"

type ConnManager struct {
	Cfg    config.AppConfig
	Logger *slog.Logger

	WSManager *WSManager
}

// OpenConnPool TODO
// In the future, it would be beneficial to open a new connection when a new request arrives.
// If a new request arrives, check the idle map:
// - If the idle map does not have a free connection, then spawn a new connection.
// Something similar to the solution in the SQL package:
//
// sql/sql.go
// Runs in a separate goroutine, opens new connections when requested.
//
//	func (db *DB) connectionOpener(ctx context.Context) {
//		 for {
//			 select {
//			 case <-ctx.Done():
//				 return
//			 case <-db.openerCh:
//				 db.openNewConnection(ctx)
//			 }
//		 }
//	}
func (cs ConnManager) OpenConnPool(userId, password string, initSize int) {
	go func() {
		ctx := logging.AppendAttrsCtx(context.Background(), logging.UserIDAttr(userId), logging.ServiceName(serviceName))
		if initSize < 1 {
			cs.Logger.WarnContext(ctx, "Connection pool initial size not specified")
		}
		for connNumber := 1; connNumber <= initSize; connNumber++ {
			ctx := logging.AppendAttrsCtx(ctx, logging.ConnNo(connNumber))
			go cs.keepUserClientConnected(ctx, userId, password, connNumber)
		}
	}()
}

func (cs ConnManager) keepUserClientConnected(ctx context.Context, userId, password string, connNumber int) {
	attempt := 1
	// log old streamSessionID to trace connections
	var oldSSID string
	for {
		// new variable to prevent adding the same key more than once
		// initial connection does not have oldStreamID
		oldSSIDCtx := logging.AppendAttrsCtx(ctx, logging.OldStreamID(oldSSID))
		wsClient, err := cs.newConnection(oldSSIDCtx, userId, password, connNumber)
		if err != nil {
			cs.Logger.WarnContext(oldSSIDCtx, "Unable to login client at the moment", logging.ErrorAttr(err),
				logging.AttemptAttr(attempt))
			if attempt == cs.Cfg.Client.Connection.MaxAttempts {
				cs.Logger.ErrorContext(oldSSIDCtx, "Critical error connection could not be established")
				// TODO
				// in future maybe send email
				panic("critical error connection could not be established check logs and XTB platform")
				return
			}
			attempt++
			time.Sleep(time.Second * time.Duration(cs.Cfg.Client.Connection.ReConnectNextTrySec))
			continue
		}
		// new variable to prevent adding the same key more than once
		newSSIDCtx := logging.AppendAttrsCtx(oldSSIDCtx, logging.StreamID(wsClient.StreamSessionID))
		go wsClient.Ping(newSSIDCtx, time.Duration(cs.Cfg.Client.Ping.IntervalSec))

		select {
		// listen for client disconnections
		case <-wsClient.ReConnCh:
			oldSSID = wsClient.StreamSessionID
			cs.Logger.InfoContext(newSSIDCtx, "Client disconnected trying to reconnect...")
		}
	}
}

// TODO do not allow further processing when at least one connection is not established
func (cs ConnManager) newConnection(ctx context.Context, userID, password string, connNumber int) (*WSClient, error) {
	cs.Logger.InfoContext(ctx, "Start establishing user connection")
	wsClient, err := cs.WSManager.DialForNewClient(ctx, cs.Cfg.XTB.Demo.WebSocketURL, nil, userID)
	if err != nil {
		return nil, fmt.Errorf("conn service create client: %w", err)
	}
	proc := processor.NewProc(wsClient)

	loginResponse, err := proc.Login(ctx, fmt.Sprintf("user=%sclient=%v", userID, connNumber),
		userID, password)
	if err != nil {
		return nil, fmt.Errorf("conn service login: %w", err)
	}
	cs.Logger.InfoContext(ctx, "Successfully established user connection", logging.RespAttr(loginResponse))

	// set client SSID
	wsClient.StreamSessionID = loginResponse.StreamSessionId
	// set client connection number
	wsClient.ConnID = connNumber
	return wsClient, nil
}
