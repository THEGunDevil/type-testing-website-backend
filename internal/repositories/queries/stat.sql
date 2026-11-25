-- name: CreateTestResult :one
INSERT INTO test_results (
    user_id,
    passage_id,
    title,
    wpm,
    raw_wpm,
    accuracy,
    duration_s,
    correct_words,
    correct_characters,
    incorrect_characters
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetUserStats :many
SELECT *
FROM test_results
WHERE user_id = $1
ORDER BY created_at DESC;