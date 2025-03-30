-- name: CreateOrganization :exec
INSERT INTO organizations (organization_id, name, create_time, create_by, last_update_time, last_update_by, row_version, is_deleted)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE organization_id = $1 LIMIT 1;

-- name: ExistOrganizationByName :one
SELECT EXISTS(SELECT 1 FROM organizations WHERE name = $1 LIMIT 1);

-- name: UpdateOrganization :exec
UPDATE organizations
SET
    name = $2,
    last_update_time = $3,
    last_update_by = $4,
    row_version = $5,
    is_deleted = $6
WHERE organization_id = $1;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE organization_id = $1;

-- name: DeleteOrganizationByName :exec
DELETE FROM organizations WHERE name = $1;

-- name: ListOrganizations :many
SELECT *
FROM organizations
WHERE
    -- Optional is_deleted filter
    sqlc.narg('is_deleted')::boolean IS NULL OR is_deleted = sqlc.narg('is_deleted')::boolean
    AND (
        -- Optional page cursor
        sqlc.narg('cursor_value')::timestamptz IS NULL -- Ignore if no cursor
        OR (sqlc.narg('is_cursor_forward')::boolean = true AND create_time > sqlc.narg('cursor_value')::timestamptz) -- Next page
        OR (sqlc.narg('is_cursor_forward')::boolean = false AND create_time < sqlc.narg('cursor_value')::timestamptz) -- Previous page
    )
ORDER BY
    -- Order depends on direction
    CASE WHEN sqlc.narg('is_cursor_forward')::boolean = true THEN create_time ELSE create_time END ASC
LIMIT CASE WHEN sqlc.narg('page_size')::int IS NULL THEN 100 ELSE sqlc.narg('page_size') END;

-- name: HasMorePagesOrganizationList :one
SELECT EXISTS(
    SELECT 1
    FROM organizations
    WHERE
        create_time > sqlc.narg('cursor_next')::timestamptz
    LIMIT 1
) AS has_next,
EXISTS(
    SELECT 1
    FROM organizations
    WHERE
        create_time < sqlc.narg('cursor_prev')::timestamptz
    ORDER BY create_time DESC
    LIMIT 1
) AS has_prev;
