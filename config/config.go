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
