package config

import (
	"github.com/artufi/trader/infrastructure/database"
	"time"
)

type AppConfig struct {
	XTB struct {
		Demo struct {
			WebSocketURL    string
			WebSocketTLSURL string

			UserID   string
			Password string
		}
	}
	Database database.PostgresConfig
	Client   struct {
		Interval time.Duration
	}
}

type Loader interface {
	Load() (AppConfig, error)
}

func MustLoad(loader Loader) AppConfig {
	cfg, err := loader.Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
