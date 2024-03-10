package processor

import (
	"context"
	"fmt"
	"github.com/artufi/trader/xtb/jsonform"
	"github.com/artufi/trader/xtb/response"
)

const logoutProc = "Logout"

func (p Proc) Logout(ctx context.Context, customTag string) (response.General, error) {
	logoutJSON, err := jsonform.Logout(customTag)
	if err != nil {
		return response.General{}, fmt.Errorf("XTB %s processor: %w", logoutProc, err)
	}

	return process[response.General](ctx, p.Client, customTag, logoutJSON, logoutProc)
}
