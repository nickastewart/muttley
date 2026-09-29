-- name: CreateMagicLink :one
INSERT INTO magic_link (email, token_hash, purpose, expires_at)
VALUES (?, ?, ?, ?)
RETURNING id, email, token_hash, purpose, expires_at, used_at, created_at;

-- name: InvalidateUnusedMagicLinks :exec
UPDATE magic_link
SET used_at = CURRENT_TIMESTAMP
WHERE email = ? AND used_at IS NULL;

-- name: GetActiveMagicLinkByTokenHash :one
SELECT id, email, token_hash, purpose, expires_at, used_at, created_at
FROM magic_link
WHERE token_hash = ? AND used_at IS NULL AND expires_at > sqlc.arg(now);

-- name: ConsumeMagicLink :execrows
UPDATE magic_link
SET used_at = CURRENT_TIMESTAMP
WHERE id = ? AND used_at IS NULL;
