package response

type GeneralResponse struct {
	Status          bool                   `json:"status"`
	StreamSessionId string                 `json:"streamSessionId"`
	ReturnData      map[string]interface{} `json:"returnData,omitempty"`
	ErrorCode       string                 `json:"errorCode,omitempty"`
	ErrorDescr      string                 `json:"errorDescr,omitempty"`
}
