CREATE TABLE IF NOT EXISTS items (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    details TEXT,
    created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    updated_on INTEGER,
    completed_on INTEGER DEFAULT NULL,
    created_by INTEGER,
    FOREIGN KEY(created_by) REFERENCES users(id)
);

