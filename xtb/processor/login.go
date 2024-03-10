package processor

import (
	"context"
	"fmt"
	"github.com/artufi/trader/xtb/command"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

const loginProc = "Login"

func (p Proc) Login(ctx context.Context, customTag, userID, password string) (response.General, error) {
	loginJSON, err := jsonform.Login(
		command.LoginArgs{
			UserID:   userID,
			Password: password,
		},
		customTag)
	if err != nil {
		return response.General{}, fmt.Errorf("XTB %s processor: %w", loginProc, err)
	}

	return process[response.General](ctx, p.Client, customTag, loginJSON, loginProc)
}
