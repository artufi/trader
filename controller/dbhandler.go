package controller

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"log/slog"
)

const (
	orderName    = "order"
	positionName = "position"
)

type DBHandler struct {
	TraceID string
	Logger  *slog.Logger

	PositionService model.PositionService
	OrderService    model.OrderService
}

func (h DBHandler) InsertOrder(ctx context.Context, order model.Order) (int, error) {
	h.Logger.InfoContext(ctx, fmt.Sprintf("Inserting %s", orderName), orderName, order)
	id, err := h.OrderService.Insert(order)
	if err != nil {
		return h.handleInsertionError(ctx, err, id, orderName)
	}
	return id, nil
}

func (h DBHandler) InsertPosition(ctx context.Context, position model.Position) (int, error) {
	h.Logger.InfoContext(ctx, fmt.Sprintf("Inserting %s", positionName), positionName, position)
	id, err := h.PositionService.Insert(position)
	if err != nil {
		return h.handleInsertionError(ctx, err, id, positionName)
	}
	return id, nil
}

func (h DBHandler) handleInsertionError(ctx context.Context, err error, id int, operation string) (int, error) {
	if id == 0 {
		h.Logger.ErrorContext(ctx, fmt.Sprintf("Failed to insert %s: no rows affected", operation),
			logging.ErrorAttr(err))
		return -1, fmt.Errorf("insert order: %w", sql.ErrNoRows)
	}
	h.Logger.WarnContext(ctx,
		fmt.Sprintf("Error occurred while inserting %s but record was inserted successfully", operation),
		logging.ErrorAttr(err),
		logging.IDAttr(id))
	return id, nil
}
