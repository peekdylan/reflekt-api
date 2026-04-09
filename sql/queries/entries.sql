-- name: CreateEntry :one
INSERT INTO entries (id, user_id, title, body, tags, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
RETURNING *;

-- name: GetEntriesByUserID :many
SELECT * FROM entries WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetEntryByID :one
SELECT * FROM entries WHERE id = $1 AND user_id = $2;

-- name: UpdateEntryAIAnalysis :one
UPDATE entries SET ai_analysis = $1, mood = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: DeleteEntry :exec
DELETE FROM entries WHERE id = $1 AND user_id = $2;