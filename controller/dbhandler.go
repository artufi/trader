package controller

import (
	"context"
	"fmt"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"log/slog"
)

const (
	orderName      = "order"
	predictionName = "prediction"
)

type DBHandler struct {
	TraceID string
	Logger  *slog.Logger

	PredictionService model.PredictionService
	OrderService      model.OrderService
}

func (h DBHandler) InsertOrder(ctx context.Context, order model.Order) (int, error) {
	h.Logger.InfoContext(ctx, fmt.Sprintf("Inserting %s", orderName), orderName, order)
	id, err := h.OrderService.Insert(order)
	if err != nil {
		return h.handleInsertionError(ctx, err, id, orderName)
	}
	return id, nil
}

func (h DBHandler) InsertPrediction(ctx context.Context, prediction model.Prediction) (int, error) {
	h.Logger.InfoContext(ctx, fmt.Sprintf("Inserting %s", predictionName), predictionName, prediction)
	id, err := h.PredictionService.Insert(prediction)
	if err != nil {
		return h.handleInsertionError(ctx, err, id, predictionName)
	}
	return id, nil
}

func (h DBHandler) handleInsertionError(ctx context.Context, err error, id int, operation string) (int, error) {
	// Auto increment on database ID, if record is inserted then id is always != 0
	if id == 0 {
		return -1, fmt.Errorf("insert no rows affected: %v", err)
	}
	h.Logger.WarnContext(ctx,
		fmt.Sprintf("Error occurred while inserting %s but record was inserted successfully", operation),
		logging.ErrorAttr(err),
		logging.IDAttr(id))
	return id, nil
}
