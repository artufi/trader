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

type ConnService struct {
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
func (cs ConnService) OpenConnPool(userId, password string, size int) {
	go func() {
		for connNumber := 1; connNumber <= size; connNumber++ {
			go cs.KeepUserClientConnected(userId, password, connNumber)
		}
	}()
}

func (cs ConnService) KeepUserClientConnected(userId, password string, connNumber int) {
	attempt := 1
	ctx := logging.AppendAttrsCtx(context.Background(), logging.ConnNo(connNumber), logging.UserIDAttr(userId))
	for {
		wsClient, err := cs.NewConnection(ctx, userId, password, connNumber)
		if err != nil {
			cs.Logger.ErrorContext(ctx, "Error while Login client", logging.ErrorAttr(err), "attempt", attempt)
			if attempt == cs.Cfg.Client.Connection.MaxAttempts {
				cs.Logger.InfoContext(ctx, "Critical error connection could not be established")
				// TODO
				// in future maybe send email
				return
			}
			attempt++
			time.Sleep(time.Second * 20)
			continue
		}
		ctx = logging.AppendAttrsCtx(ctx, logging.StreamID(wsClient.StreamSessionID))
		go wsClient.Ping(ctx, time.Duration(cs.Cfg.Client.Ping.IntervalSec))
		select {
		// listen for client disconnections
		case <-wsClient.ReConnCh:
			cs.Logger.InfoContext(ctx, "Client disconnected trying to reconnect...")
		}
	}
}

func (cs ConnService) NewConnection(ctx context.Context, userID, password string, connNumber int) (*WSClient, error) {
	cs.Logger.InfoContext(ctx, "Start establishing user connection")
	wsClient, err := cs.WSManager.DialForNewClient(ctx, cs.Cfg.XTB.Demo.WebSocketURL, nil, userID)
	if err != nil {
		cs.Logger.ErrorContext(ctx, "Failed to create a new client", logging.ErrorAttr(err))
		return nil, fmt.Errorf("conn service: %w", err)
	}
	proc := processor.NewProc(wsClient)

	loginResponse, err := proc.Login(ctx, fmt.Sprintf("user=%sclient=%v", userID, connNumber),
		userID, password)
	if err != nil {
		cs.Logger.ErrorContext(ctx, "Failed to login", logging.ErrorAttr(err))
		return nil, fmt.Errorf("conn service: %w", err)
	}
	cs.Logger.InfoContext(ctx, "Successfully established user connection", logging.RespAttr(loginResponse))

	// set client SSID
	wsClient.StreamSessionID = loginResponse.StreamSessionId
	return wsClient, nil
}
