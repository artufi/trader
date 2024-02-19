package command

type GetAllSymbols struct {
	Command string `json:"command"`
}

type GetSymbol struct {
	Command   string        `json:"command"`
	Arguments GetSymbolArgs `json:"arguments"`
}

type GetSymbolArgs struct {
	Symbol string `json:"symbol"`
}
