package response

import "fmt"

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
		return fmt.Errorf("XTB response General status: %t, errorCode: %s, errorDescr: %s",
			g.Status, g.ErrorCode, g.ErrorDescr)
	}
	return nil
}

func (g General) CheckCustomTag(customTag string) error {
	if g.CustomTag != customTag {
		return fmt.Errorf("XTB response General customTag is not equal, expected: %s, got: %s",
			customTag, g.CustomTag)
	}
	return nil
}
