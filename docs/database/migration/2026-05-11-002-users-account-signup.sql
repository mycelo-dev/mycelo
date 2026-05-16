CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    tenant_id BIGINT REFERENCES tenants(tenant_id),
    user_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(500) NOT NULL,
    created_at BIGINT,
    updated_at BIGINT,

    CONSTRAINT uq_users_email UNIQUE(email)
);

CREATE INDEX IF NOT EXISTS idx_users_tenant_public
ON users (tenant_id, user_public_id);

ALTER TABLE teams
DROP CONSTRAINT IF EXISTS uq_team_name_per_tenant;

ALTER TABLE teams
ADD CONSTRAINT uq_team_name_per_tenant UNIQUE(tenant_id, team_name);
