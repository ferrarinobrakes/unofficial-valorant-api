-- accounts table: cache for resolved account data
CREATE TABLE IF NOT EXISTS accounts (
    puuid TEXT PRIMARY KEY,
    region TEXT NOT NULL,
    account_level INTEGER NOT NULL,
    name TEXT NOT NULL,
    tag TEXT NOT NULL,
    card TEXT NOT NULL,
    title TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_accounts_name_tag ON accounts(name, tag);
CREATE INDEX IF NOT EXISTS idx_accounts_updated_at ON accounts(updated_at);

-- clients table: connected client nodes
CREATE TABLE IF NOT EXISTS clients (
    client_id TEXT PRIMARY KEY,
    last_heartbeat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    lcu_available BOOLEAN NOT NULL DEFAULT FALSE,
    connected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_clients_last_heartbeat ON clients(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_clients_lcu_available ON clients(lcu_available);

-- request_log table: audit trail for requests
CREATE TABLE IF NOT EXISTS request_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id TEXT NOT NULL,
    client_id TEXT,
    game_name TEXT NOT NULL,
    game_tag TEXT NOT NULL,
    puuid TEXT,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_request_log_request_id ON request_log(request_id);
CREATE INDEX IF NOT EXISTS idx_request_log_created_at ON request_log(created_at);
CREATE INDEX IF NOT EXISTS idx_request_log_success ON request_log(success);
