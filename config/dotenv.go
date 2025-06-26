package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type DotEnvCfg struct {
	Filename string
}

func (dec DotEnvCfg) Load() (AppConfig, error) {
	cfg := AppConfig{}

	err := godotenv.Load(dec.Filename)
	if err != nil {
		return cfg, fmt.Errorf("load dotenv file=%s: %w", dec.Filename, err)
	}

	cfg.XTB.Demo.WebSocketURL = os.Getenv("WEBSOCKET_URL")
	cfg.XTB.Demo.WebSocketStreamURL = os.Getenv("WEBSOCKET_STREAM_URL")

	cfg.XTB.Demo.UserID = os.Getenv("USER_ID")
	cfg.XTB.Demo.Password = os.Getenv("USER_PASSWORD")

	cfg.Database.Host = os.Getenv("PSQL_HOST")
	cfg.Database.Port = os.Getenv("PSQL_PORT")
	cfg.Database.User = os.Getenv("PSQL_USER")
	cfg.Database.Password = os.Getenv("PSQL_PASSWORD")
	cfg.Database.Database = os.Getenv("PSQL_DATABASE")
	cfg.Database.SSLMode = os.Getenv("PSQL_SSLMODE")
	cfg.Database.Connection.MaxAttempts, err = strconv.Atoi(os.Getenv("DB_CONNECTION_MAX_ATTEMPTS"))
	if err != nil {
		panic(err)
	}
	cfg.Database.Connection.NextTrySec, err = strconv.Atoi(os.Getenv("DB_CONNECTION_NEXT_TRY_TIME_SEC"))
	if err != nil {
		panic(err)
	}

	cfg.Client.Ping.IntervalSec, err = strconv.Atoi(os.Getenv("WS_CLIENT_PING_INTERVAL_SEC"))
	if err != nil {
		panic(err)
	}
	cfg.Client.Connection.Pool.Size, err = strconv.Atoi(os.Getenv("WS_CLIENT_CONNECTION_POOL_SIZE"))
	if err != nil {
		panic(err)
	}
	cfg.Client.Connection.MaxAttempts, err = strconv.Atoi(os.Getenv("WS_CLIENT_CONNECTION_MAX_ATTEMPTS"))
	if err != nil {
		panic(err)
	}
	cfg.Client.Connection.ReConnectNextTrySec, err = strconv.Atoi(os.Getenv("WS_CLIENT_CONNECTION_RECONNECT_NEXT_TRY_TIME_SEC"))
	if err != nil {
		panic(err)
	}

	cfg.Client.Stream.Ping.IntervalSec, err = strconv.Atoi(os.Getenv("WS_CLIENT_STREAM_PING_INTERVAL_SEC"))
	if err != nil {
		panic(err)
	}

	cfg.Client.Order.Close.IntervalSec, err = strconv.Atoi(os.Getenv("WS_CLIENT_ORDER_CLOSE_INTERVAL_SEC"))
	if err != nil {
		panic(err)
	}

	cfg.BuySellParams.Volume, err = strconv.ParseFloat(os.Getenv("VOLUME_TO_BUY_SELL"), 64)
	if err != nil {
		panic(err)
	}

	return cfg, nil
}
