package processor

import (
	"context"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

const loginProc = "login"

func (p Proc) Login(ctx context.Context, customTag, userID, password string) (response.General, error) {
	loginJSON, err := jsonform.Login(
		command.LoginArgs{
			UserID:   userID,
			Password: password,
		},
		customTag)
	if err != nil {
		return response.General{}, &ProcError{Processor: loginProc, Err: err}
	}

	return process[response.General](ctx, p.client, customTag, loginJSON, loginProc)
}
