CREATE TABLE IF NOT EXISTS oauth_clients (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    scopes TEXT NOT NULL,
    created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    updated_on INTEGER,
    archived_on INTEGER DEFAULT NULL
);
