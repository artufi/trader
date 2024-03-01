package response

import "fmt"

type General struct {
	Status          bool                   `json:"status"`
	StreamSessionId string                 `json:"streamSessionId"`
	ReturnData      map[string]interface{} `json:"returnData,omitempty"`
	ErrorCode       string                 `json:"errorCode,omitempty"`
	ErrorDescr      string                 `json:"errorDescr,omitempty"`
}

func (gr General) CheckStatus() error {
	if !gr.Status {
		return fmt.Errorf("XTB General response status: %t, errorCode: %s, errorDescr: %s",
			gr.Status, gr.ErrorCode, gr.ErrorDescr)
	}
	return nil
}
