package command

type Ping struct {
	Command         string `json:"command"`
	CustomTag       string `json:"customTag,omitempty"`
	StreamSessionId string `json:"streamSessionId,omitempty"`
}
