-- name: LogRequest :exec
INSERT INTO request_log (request_id, client_id, game_name, game_tag, puuid, success, error_message, duration_ms)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetRequestStats :one
SELECT
    COUNT(*) as total_requests,
    SUM(CASE WHEN success = TRUE THEN 1 ELSE 0 END) as successful_requests,
    AVG(duration_ms) as avg_duration_ms
FROM request_log
WHERE created_at > datetime('now', '-1 hour');

-- name: CleanOldLogs :exec
DELETE FROM request_log
WHERE created_at < datetime('now', '-30 days');
