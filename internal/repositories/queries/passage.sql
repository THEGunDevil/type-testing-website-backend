-- name: CreatePassage :one
INSERT INTO passages (
    content,
    language
)
VALUES ($1, $2)
RETURNING *;
-- name: GetRandomPassage :one
SELECT *
FROM passages
WHERE language = $1
ORDER BY RANDOM()
LIMIT 1;