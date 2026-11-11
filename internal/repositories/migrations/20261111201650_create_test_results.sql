-- +goose Up
-- +goose StatementBegin
CREATE TABLE test_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    passage_id UUID REFERENCES passages(id) ON DELETE SET NULL,
    wpm REAL NOT NULL,
    accuracy REAL NOT NULL CHECK (accuracy BETWEEN 0 AND 100),
    duration_s INT NOT NULL,
    mistakes INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_results_user ON test_results (user_id, created_at DESC);

CREATE VIEW leaderboard AS
SELECT u.id AS user_id,
       u.user_name,
       MAX(r.wpm) AS best_wpm,
       AVG(r.accuracy) AS avg_accuracy
FROM test_results r
JOIN users u ON u.id = r.user_id
GROUP BY u.id, u.user_name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS leaderboard;
DROP TABLE IF EXISTS test_results;
-- +goose StatementEnd