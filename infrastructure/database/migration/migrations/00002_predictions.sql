-- +goose Up
-- +goose StatementBegin
CREATE TABLE predictions
(
    id                  SERIAL PRIMARY KEY,
    user_id             INT,
    symbol              TEXT,
    prediction_date     TEXT,
    allocation          DOUBLE PRECISION,
    preds_proba         DOUBLE PRECISION,
    stop_loss           DOUBLE PRECISION,
    take_profit         DOUBLE PRECISION,
    forecast_horizon    INT,
    target_change       INT,
    model_type          TEXT,
    run_timestamp       TEXT,
    sl_tp_logic         TEXT,
    id_model_properties TEXT,
    interval            TEXT,
    created_at          TIMESTAMPTZ,
    CONSTRAINT fk_predictions_user FOREIGN KEY (user_id) REFERENCES users (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP table predictions;
-- +goose StatementEnd