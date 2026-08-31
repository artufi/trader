package response

type TradeTransaction struct {
	Status          bool   `json:"status"`
	StreamSessionId string `json:"streamSessionId,omitempty"`
	ReturnData      struct {
		Order int `json:"order"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
	CustomTag  string `json:"customTag"`
}

func (tt TradeTransaction) CheckStatus() error {
	if !tt.Status {
		return &StatusError{
			Status:     false,
			ErrorCode:  tt.ErrorCode,
			ErrorDescr: tt.ErrorDescr,
		}
	}
	return nil
}

func (tt TradeTransaction) CheckCustomTag(customTag string) error {
	if tt.CustomTag != customTag {
		return &CustomTagError{
			Returned: tt.CustomTag,
			Provided: customTag,
		}
	}
	return nil
}

type TradeTransactionStatus struct {
	Status          bool   `json:"status"`
	StreamSessionId string `json:"streamSessionId,omitempty"`
	ReturnData      struct {
		Ask           float64       `json:"ask"`
		Bid           float64       `json:"bid"`
		CustomComment string        `json:"customComment"`
		Message       *string       `json:"message"`
		Order         int           `json:"order"` // not sure maybe float64
		RequestStatus RequestStatus `json:"requestStatus"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
	CustomTag  string `json:"customTag"`
}

type RequestStatus int

const (
	ERROR    RequestStatus = 0
	PENDING  RequestStatus = 1
	ACCEPTED RequestStatus = 3
	REJECTED RequestStatus = 4
)

var RequestStatusName = map[RequestStatus]string{
	ERROR:    "ERROR",
	PENDING:  "PENDING",
	ACCEPTED: "ACCEPTED",
	REJECTED: "REJECTED",
}

func (tts TradeTransactionStatus) CheckStatus() error {
	if !tts.Status {
		return &StatusError{
			Status:     false,
			ErrorCode:  tts.ErrorCode,
			ErrorDescr: tts.ErrorDescr,
		}
	}
	return nil
}

func (tts TradeTransactionStatus) CheckCustomTag(customTag string) error {
	if tts.CustomTag != customTag {
		return &CustomTagError{
			Returned: tts.CustomTag,
			Provided: customTag,
		}
	}
	return nil
}

func (tts TradeTransactionStatus) CheckRequestStatus() (RequestStatus, error) {
	switch tts.ReturnData.RequestStatus {
	case ACCEPTED, PENDING:
		return tts.ReturnData.RequestStatus, nil
	default:
		return tts.newRequestStatusError()
	}
}

func (tts TradeTransactionStatus) newRequestStatusError() (RequestStatus, error) {
	status, ok := RequestStatusName[tts.ReturnData.RequestStatus]
	if !ok {
		status = "UNKNOWN"
	}
	rse := &RequestStatusError{
		RequestStatus: status,
	}
	if tts.ReturnData.Message != nil {
		rse.Message = *tts.ReturnData.Message
	}
	return tts.ReturnData.RequestStatus, rse
}
