CREATE TABLE IF NOT EXISTS items (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    details TEXT,
    created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    completed_on INTEGER
);
