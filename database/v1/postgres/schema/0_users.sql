CREATE TABLE IF NOT EXISTS users (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "username" text NOT NULL,
    "hashed_password" text NOT NULL,
    "password_last_changed_on" timestamp,
    "two_factor_secret" text NOT NULL,
    "is_admin" bool NOT NULL DEFAULT 'false',
    "created_on" timestamp NOT NULL DEFAULT to_timestamp(extract(epoch FROM NOW())),
    "updated_on" timestamp,
    "archived_on" timestamp,
    UNIQUE ("username")
);
