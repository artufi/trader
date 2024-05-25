package response

import "fmt"

type StatusError struct {
	Status     bool
	ErrorCode  string
	ErrorDescr string
}

func (se *StatusError) Error() string {
	return fmt.Sprintf("response status: %t, errorCode: %s, errorDescr: %s",
		se.Status, se.ErrorCode, se.ErrorDescr)
}

type CustomTagError struct {
	Returned string
	Provided string
}

func (cte *CustomTagError) Error() string {
	return fmt.Sprintf("response customTag, returned: %s, provided: %s", cte.Returned, cte.Provided)
}

type RequestStatusError struct {
	RequestStatus string
	Message       string
}

func (rse *RequestStatusError) Error() string {
	return fmt.Sprintf("response requestStatus: %s, message: %s", rse.RequestStatus, rse.Message)
}
