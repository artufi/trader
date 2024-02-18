package config

type Config struct {
	XTB struct {
		Demo struct {
			WebSocketURL    string `yaml:"websocketURL"`
			WebSocketTLSURL string `yaml:"websocketTLSURL"`
		} `yaml:"demo"`
	} `yaml:"XTB"`
	Testing struct {
		UserID   string `json:"userID"`
		Password string `json:"password"`
	} `json:"testing"`
}
