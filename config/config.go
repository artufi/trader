package config

import (
	"github.com/artufi/trader/infrastructure/database"
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

	Database struct {
		database.PostgresConfig
		Connection struct {
			MaxAttempts int
			NextTrySec  int
		}
	}

	Client struct {
		Ping struct {
			IntervalSec int
		}
		Connection struct {
			Pool struct {
				Size int
			}
			MaxAttempts int
		}
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
