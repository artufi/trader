package response

type General struct {
	Status          bool                   `json:"status"`
	StreamSessionId string                 `json:"streamSessionId,omitempty"`
	ReturnData      map[string]interface{} `json:"returnData,omitempty"`
	ErrorCode       string                 `json:"errorCode,omitempty"`
	ErrorDescr      string                 `json:"errorDescr,omitempty"`
	CustomTag       string                 `json:"customTag"`
}

func (g General) CheckStatus() error {
	if !g.Status {
		return &StatusError{
			Status:     false,
			ErrorCode:  g.ErrorCode,
			ErrorDescr: g.ErrorDescr,
		}
	}
	return nil
}

func (g General) CheckCustomTag(customTag string) error {
	if g.CustomTag != customTag {
		return &CustomTagError{
			Returned: g.CustomTag,
			Provided: customTag,
		}
	}
	return nil
}
