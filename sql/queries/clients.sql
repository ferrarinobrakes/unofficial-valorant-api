-- name: RegisterClient :exec
INSERT INTO clients (client_id, version, last_heartbeat, lcu_available, connected_at)
VALUES (?, ?, CURRENT_TIMESTAMP, ?, CURRENT_TIMESTAMP)
ON CONFLICT(client_id) DO UPDATE SET
    version = excluded.version,
    last_heartbeat = CURRENT_TIMESTAMP,
    lcu_available = excluded.lcu_available,
    connected_at = CURRENT_TIMESTAMP;

-- name: UpdateClientHeartbeat :exec
UPDATE clients
SET last_heartbeat = CURRENT_TIMESTAMP,
    lcu_available = ?
WHERE client_id = ?;

-- name: GetAvailableClients :many
SELECT * FROM clients
WHERE lcu_available = TRUE
  AND last_heartbeat > datetime('now', '-30 seconds')
ORDER BY last_heartbeat DESC;

-- name: GetAllClients :many
SELECT * FROM clients
ORDER BY last_heartbeat DESC;

-- name: RemoveClient :exec
DELETE FROM clients
WHERE client_id = ?;

-- name: CleanStaleClients :exec
DELETE FROM clients
WHERE last_heartbeat < datetime('now', '-2 minutes');
