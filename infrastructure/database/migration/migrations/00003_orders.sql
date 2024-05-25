-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders
(
    id                 SERIAL PRIMARY KEY,
    order_no           BIGINT NOT NULL,
    position           BIGINT,
    user_id            INT,
    prediction_id      INT,
    symbol             TEXT,
    operation_code_cmd INT,
    transaction_type   INT,
    custom_comment     TEXT,
    expiration         BIGINT,
    offst              INT,
    ordr               INT,
    price              FLOAT,
    stop_loss          FLOAT,
    take_profit        FLOAT,
    volume             FLOAT,
    tts_request_status TEXT,
    tts_message        TEXT,
    open_price         FLOAT,
    close_price        FLOAT,
    profit             FLOAT,
    closed             BOOLEAN,
    comment            TEXT,
    close_time         TIMESTAMPTZ,
    open_time          TIMESTAMPTZ,
    created_at         TIMESTAMPTZ,
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_orders_prediction FOREIGN KEY (prediction_id) REFERENCES predictions (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd