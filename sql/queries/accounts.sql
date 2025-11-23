-- name: GetAccountByNameTag :one
SELECT * FROM accounts
WHERE name = ? AND tag = ?
LIMIT 1;

-- name: GetAccountByPUUID :one
SELECT * FROM accounts
WHERE puuid = ?
LIMIT 1;

-- name: UpsertAccount :exec
INSERT INTO accounts (puuid, region, account_level, name, tag, card, title, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(puuid) DO UPDATE SET
    region = excluded.region,
    account_level = excluded.account_level,
    name = excluded.name,
    tag = excluded.tag,
    card = excluded.card,
    title = excluded.title,
    updated_at = CURRENT_TIMESTAMP;

-- name: CleanOldAccounts :exec
DELETE FROM accounts
WHERE updated_at < datetime('now', '-7 days');
