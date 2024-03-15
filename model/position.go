package model

import (
	"database/sql"
	"fmt"
	"time"
)

type PositionService struct {
	DB *sql.DB
}

type Position struct {
	ID      int
	UserID  int
	OrderID int
	PurchaseInstruction
}

func (ps PositionService) Insert(p Position) (int, error) {
	row := ps.DB.QueryRow(`
		INSERT INTO position
			(user_id,
			 order_id,
			 symbol, 
			 date,
			 allocation,
			 preds_proba,
			 stop_loss,
			 take_profit,
			 forecast_horizon,
			 target_change,
			 model_type,
			 created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`, p.UserID, p.OrderID, p.Symbol, p.Date, p.Allocation, p.PredsProba, p.StopLoss, p.TakeProfit,
		p.ModelSetup.ForecastHorizon, p.ModelSetup.TargetChange, p.ModelSetup.ModelType, time.Now())

	err := row.Scan(&p.ID)
	if err != nil {
		return -1, fmt.Errorf("insert position: %w", err)
	}
	return p.ID, nil
}
