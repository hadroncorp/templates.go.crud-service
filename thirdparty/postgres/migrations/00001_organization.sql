-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organizations (
    organization_id VARCHAR(48) PRIMARY KEY,
    name TEXT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL,
    create_by VARCHAR(96) NOT NULL ,
    last_update_time TIMESTAMPTZ NOT NULL,
    last_update_by VARCHAR(96) NOT NULL,
    row_version BIGINT NOT NULL,
    is_deleted BOOLEAN NOT NULL
);
-- For name-based searches
CREATE INDEX idx_organizations_name ON organizations(name, organization_id) WHERE is_deleted = false;
-- For time-based pagination
CREATE INDEX idx_organizations_create_time ON organizations(create_time DESC, organization_id DESC) WHERE is_deleted = false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_organizations_create_time;
DROP INDEX IF EXISTS idx_organizations_name;
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd
