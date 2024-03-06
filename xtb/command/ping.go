package command

type Ping struct {
	Command   string `json:"command"`
	CustomTag string `json:"customTag"`
}
