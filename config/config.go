package config

type AppConfig struct {
	XTB struct {
		Demo struct {
			WebSocketURL    string `yaml:"websocketURL"`
			WebSocketTLSURL string `yaml:"websocketTLSURL"`

			UserID   string `yaml:"userID"`
			Password string `yaml:"password"`
		} `yaml:"demo"`
	} `yaml:"XTB"`
}

func MustLoad(cfg AppConfig, err error) AppConfig {
	if err != nil {
		panic(err)
	}
	return cfg
}
