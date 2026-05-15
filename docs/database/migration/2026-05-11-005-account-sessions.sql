CREATE TABLE IF NOT EXISTS account_sessions (
    session_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    session_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    tenant_public_id UUID NOT NULL REFERENCES tenants(tenant_public_id),
    user_public_id UUID NOT NULL REFERENCES users(user_public_id),
    session_hash VARCHAR(500) NOT NULL UNIQUE,
    created_at BIGINT,
    expires_at BIGINT
);

CREATE INDEX IF NOT EXISTS idx_account_sessions_hash_expires
ON account_sessions (session_hash, expires_at);
