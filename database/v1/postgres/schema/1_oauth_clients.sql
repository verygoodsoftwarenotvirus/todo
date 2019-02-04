CREATE TABLE IF NOT EXISTS oauth_clients (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "client_id" TEXT NOT NULL,
    "client_secret" TEXT NOT NULL,
    "redirect_uri" TEXT DEFAULT '',
    "scopes" TEXT NOT NULL,
    "implicit_allowed" BOOLEAN NOT NULL DEFAULT 'false',
    "created_on" timestamp NOT NULL DEFAULT to_timestamp(extract(epoch FROM NOW())),
    "updated_on" timestamp DEFAULT NULL,
    "archived_on" timestamp DEFAULT NULL,
    "belongs_to" INTEGER NOT NULL,
    FOREIGN KEY(belongs_to) REFERENCES users(id)
);
