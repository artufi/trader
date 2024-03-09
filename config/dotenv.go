package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

func LoadDotEnvConfig(filename string) (AppConfig, error) {
	cfg := AppConfig{}

	err := godotenv.Load(filename)
	if err != nil {
		return cfg, fmt.Errorf("load .env file: %w", err)
	}

	cfg.XTB.Demo.WebSocketURL = os.Getenv("WEBSOCKET_URL")

	cfg.XTB.Demo.UserID = os.Getenv("USER_ID")
	cfg.XTB.Demo.Password = os.Getenv("USER_PASSWORD")

	return cfg, nil
}

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

	cfg.XTB.Demo.UserID = os.Getenv("USER_ID")
	cfg.XTB.Demo.Password = os.Getenv("USER_PASSWORD")

	return cfg, nil
}
