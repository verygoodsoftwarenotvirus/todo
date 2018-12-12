CREATE TABLE IF NOT EXISTS users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    password_last_changed_on INTEGER,
    two_factor_secret TEXT NOT NULL,
    created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    updated_on INTEGER,
    archived_on INTEGER DEFAULT NULL,
    CONSTRAINT username_unique UNIQUE (username)
);
