package config

import "github.com/artufi/trader/infrastructure/database"

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
