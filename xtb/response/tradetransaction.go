package response

import "fmt"

type TradeTransaction struct {
	Status          bool   `json:"status"`
	StreamSessionId string `json:"streamSessionId"`
	ReturnData      struct {
		Order int `json:"order"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
}

func (tt TradeTransaction) CheckStatus() error {
	if !tt.Status {
		return fmt.Errorf("XTB response TradeTransaction status: %t, errorCode: %s, errorDescr: %s",
			tt.Status, tt.ErrorCode, tt.ErrorDescr)
	}
	return nil
}

type TradeTransactionStatus struct {
	Status          bool   `json:"status"`
	StreamSessionId string `json:"streamSessionId"`
	ReturnData      struct {
		Ask           float64       `json:"ask"`
		Bid           float64       `json:"bid"`
		CustomComment string        `json:"customComment"`
		Message       interface{}   `json:"message"` // not sure maybe pointer?
		Order         int           `json:"order"`   // not sure maybe float64
		RequestStatus requestStatus `json:"requestStatus"`
	} `json:"returnData"`
	ErrorCode  string `json:"errorCode,omitempty"`
	ErrorDescr string `json:"errorDescr,omitempty"`
}

type requestStatus int

const (
	ERROR requestStatus = iota
	PENDING
	ACCEPTED
	REJECTED
)

var requestStatusName = map[requestStatus]string{
	ERROR:    "ERROR",
	PENDING:  "PENDING",
	ACCEPTED: "ACCEPTED",
	REJECTED: "REJECTED",
}

func (tts TradeTransactionStatus) CheckStatus() error {
	if !tts.Status {
		return fmt.Errorf("XTB TradeTransactionStatus response requestStatus: %t, errorCode: %s, errorDescr: %s",
			tts.Status, tts.ErrorCode, tts.ErrorDescr)
	}
	return nil
}

func (tts TradeTransactionStatus) CheckRequestStatus() error {
	if tts.ReturnData.RequestStatus == ERROR || tts.ReturnData.RequestStatus == REJECTED {
		return fmt.Errorf("XTB TradeTransactionStatus response requestStatus: %d, message: %s",
			requestStatusName[tts.ReturnData.RequestStatus], tts.ReturnData.Message)
	}
	return nil
}
