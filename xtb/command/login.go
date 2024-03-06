package command

type Login struct {
	Command   string    `json:"command"`
	Arguments LoginArgs `json:"arguments"`
	CustomTag string    `json:"customTag"`
}

type LoginArgs struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
}
