-- name: CreateFamilyLink :one
INSERT INTO family_links (child_id, parent_id)
VALUES (sqlc.arg(child_id), sqlc.arg(parent_id))
RETURNING *;
