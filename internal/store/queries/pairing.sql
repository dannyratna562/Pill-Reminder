-- name: CreatePairingCode :one
INSERT INTO pairing_codes (code, parent_id, expires_at)
VALUES (sqlc.arg(code), sqlc.arg(parent_id), sqlc.arg(expires_at))
RETURNING *;

-- name: RedeemPairingCode :one
UPDATE pairing_codes
SET used_at = now()
WHERE code = sqlc.arg(code) AND used_at IS NULL AND expires_at > now()
RETURNING *;

-- name: GetPairingCodeByCode :one
SELECT * FROM pairing_codes WHERE code = sqlc.arg(code);
