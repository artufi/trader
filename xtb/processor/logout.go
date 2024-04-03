package processor

import (
	"context"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

const logoutProc = "logout"

func (p Proc) Logout(ctx context.Context, customTag string) (response.General, error) {
	logoutJSON, err := jsonform.Logout(customTag)
	if err != nil {
		return response.General{}, &ProcError{Processor: logoutProc, Err: err}
	}

	return process[response.General](ctx, p.client, customTag, logoutJSON, logoutProc)
}
