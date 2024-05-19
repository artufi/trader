-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id       SERIAL PRIMARY KEY,
    username INT UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd