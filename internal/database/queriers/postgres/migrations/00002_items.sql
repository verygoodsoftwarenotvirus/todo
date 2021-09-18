CREATE TABLE IF NOT EXISTS items (
     id CHAR(27) NOT NULL PRIMARY KEY,
     name TEXT NOT NULL,
     details TEXT NOT NULL DEFAULT '',
     created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
     last_updated_on BIGINT DEFAULT NULL,
     archived_on BIGINT DEFAULT NULL,
     belongs_to_account CHAR(27) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
);
