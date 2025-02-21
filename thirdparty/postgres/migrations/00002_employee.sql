-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employees(
    employee_id VARCHAR(48) PRIMARY KEY,
    full_name TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE employees CASCADE;
-- +goose StatementEnd
