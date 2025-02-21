-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS platform_users(
    user_id VARCHAR(48) PRIMARY KEY,
    full_name TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE platform_users CASCADE;
-- +goose StatementEnd
