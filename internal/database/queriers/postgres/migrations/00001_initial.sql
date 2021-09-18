CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL,
    created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW())
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE IF NOT EXISTS users (
    id CHAR(27) NOT NULL PRIMARY KEY,
    username TEXT NOT NULL,
    avatar_src TEXT,
    hashed_password TEXT NOT NULL,
    password_last_changed_on INTEGER,
    requires_password_change BOOLEAN NOT NULL DEFAULT 'false',
    two_factor_secret TEXT NOT NULL,
    two_factor_secret_verified_on BIGINT DEFAULT NULL,
    service_roles TEXT NOT NULL DEFAULT 'service_user',
    reputation TEXT NOT NULL DEFAULT 'unverified',
    reputation_explanation TEXT NOT NULL DEFAULT '',
    created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
    last_updated_on BIGINT DEFAULT NULL,
    archived_on BIGINT DEFAULT NULL,
    UNIQUE("username")
);

CREATE TABLE IF NOT EXISTS accounts (
    id CHAR(27) NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    billing_status TEXT NOT NULL DEFAULT 'unpaid',
    contact_email TEXT NOT NULL DEFAULT '',
    contact_phone TEXT NOT NULL DEFAULT '',
    payment_processor_customer_id TEXT NOT NULL DEFAULT '',
    subscription_plan_id TEXT,
    created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
    last_updated_on BIGINT DEFAULT NULL,
    archived_on BIGINT DEFAULT NULL,
    belongs_to_user CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE("belongs_to_user", "name")
);

CREATE TABLE IF NOT EXISTS account_user_memberships (
    id CHAR(27) NOT NULL PRIMARY KEY,
    belongs_to_account CHAR(27) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    belongs_to_user CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    default_account BOOLEAN NOT NULL DEFAULT 'false',
    account_roles TEXT NOT NULL DEFAULT 'account_user',
    created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
    last_updated_on BIGINT DEFAULT NULL,
    archived_on BIGINT DEFAULT NULL,
    UNIQUE("belongs_to_account", "belongs_to_user")
);

CREATE TABLE IF NOT EXISTS api_clients (
    id CHAR(27) NOT NULL PRIMARY KEY,
    name TEXT DEFAULT '',
    client_id TEXT NOT NULL,
    secret_key BYTEA NOT NULL,
    permissions BIGINT NOT NULL DEFAULT 0,
    admin_permissions BIGINT NOT NULL DEFAULT 0,
    created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
    last_updated_on BIGINT DEFAULT NULL,
    archived_on BIGINT DEFAULT NULL,
    belongs_to_user CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS webhooks (
    id CHAR(27) NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    url TEXT NOT NULL,
    method TEXT NOT NULL,
    events TEXT NOT NULL,
    data_types TEXT NOT NULL,
    topics TEXT NOT NULL,
    created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
    last_updated_on BIGINT DEFAULT NULL,
    archived_on BIGINT DEFAULT NULL,
    belongs_to_account CHAR(27) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
);