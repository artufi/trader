package config

type Config struct {
	XTB struct {
		Demo struct {
			WebSocketURL    string `yaml:"websocketURL"`
			WebSocketTLSURL string `yaml:"websocketTLSURL"`
		} `yaml:"demo"`
	} `yaml:"XTB"`
	Testing struct {
		UserID   string `yaml:"userID"`
		Password string `yaml:"password"`
	} `yaml:"testing"`
}
