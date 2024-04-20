package model

import (
	"database/sql"
	"fmt"
	"time"
)

type Prediction struct {
	ID     int
	UserID int
	PredictionDetails
}

type PredictionService struct {
	DB *sql.DB
}

func (ps PredictionService) Insert(p Prediction) (int, error) {
	row := ps.DB.QueryRow(`
		INSERT INTO predictions
			(user_id,
			 symbol, 
			 prediction_date,
			 allocation,
			 preds_proba,
			 stop_loss,
			 take_profit,
			 forecast_horizon,
			 target_change,
			 model_type,
			 run_timestamp,
			 sl_tp_logic,
			 id_model_properties,
			 created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`, p.UserID, p.Symbol, p.PredictionDate, p.Allocation, p.PredsProba, p.StopLoss, p.TakeProfit,
		p.ModelSetup.ForecastHorizon, p.ModelSetup.TargetChange, p.ModelSetup.ModelType, p.RunTimestamp,
		p.SlTpLogic, p.IdModelProperties, time.Now())

	err := row.Scan(&p.ID)
	if err != nil {
		return p.ID, fmt.Errorf("insert prediction: %w", err)
	}
	return p.ID, nil
}
