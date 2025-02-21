-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS places(
    place_id VARCHAR(48) PRIMARY KEY,
    display_name TEXT NOT NULL,
    location GEOGRAPHY(POINT, 4326)  -- SRID 4326 = WGS 84 (standard GPS)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE platform_users CASCADE;
DROP EXTENSION IF EXISTS postgis CASCADE;
-- +goose StatementEnd
