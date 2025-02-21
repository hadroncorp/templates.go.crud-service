-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS appointments(
    appointment_id VARCHAR(48) PRIMARY KEY,
    title TEXT NOT NULL,
    place_id VARCHAR(48) NOT NULL,
    targeted_to VARCHAR(48) DEFAULT NULL,
    scheduled_by VARCHAR(48) NOT NULL,
    scheduled_time TIMESTAMPTZ NOT NULL,
    notes TEXT DEFAULT NULL,
    status_type VARCHAR(32) NOT NULL,
    create_time TIMESTAMPTZ NOT NULL,
    create_by VARCHAR(96) NOT NULL ,
    last_update_time TIMESTAMPTZ NOT NULL,
    last_update_by VARCHAR(96) NOT NULL,
    row_version BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL,
    FOREIGN KEY (place_id) REFERENCES places(place_id) ON DELETE CASCADE,
    FOREIGN KEY (targeted_to) REFERENCES employees(employee_id) ON DELETE SET NULL,
    FOREIGN KEY (scheduled_by) REFERENCES platform_users(user_id) ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE appointments CASCADE;
-- +goose StatementEnd
