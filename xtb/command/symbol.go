package command

type GetAllSymbols struct {
	Command   string `json:"command"`
	CustomTag string `json:"customTag"`
}

type GetSymbol struct {
	Command   string        `json:"command"`
	Arguments GetSymbolArgs `json:"arguments"`
	CustomTag string        `json:"customTag"`
}

type GetSymbolArgs struct {
	Symbol string `json:"symbol"`
}
