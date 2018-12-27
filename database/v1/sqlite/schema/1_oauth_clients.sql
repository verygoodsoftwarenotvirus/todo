CREATE TABLE IF NOT EXISTS oauth_clients (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    redirect_uri TEXT DEFAULT '',
    scopes TEXT NOT NULL,
    implicit_allowed BOOLEAN NOT NULL DEFAULT 'false',
    created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    updated_on INTEGER,
    archived_on INTEGER DEFAULT NULL,
    belongs_to INTEGER NOT NULL,
    FOREIGN KEY(belongs_to) REFERENCES users(id)
);
