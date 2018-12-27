CREATE TABLE IF NOT EXISTS items (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    details TEXT,
    created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    updated_on INTEGER,
    completed_on INTEGER DEFAULT NULL,
    belongs_to INTEGER DEFAULT NULL, -- TODO: NOT NULL
    FOREIGN KEY(belongs_to) REFERENCES users(id)
);

