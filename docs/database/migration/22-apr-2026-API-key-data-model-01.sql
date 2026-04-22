CREATE TABLE api_keys (
    key_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    key_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    tenant_id BIGINT REFERENCES tenants(tenant_id),
    team_id BIGINT REFERENCES teams(team_id),
    key_hash VARCHAR(500),
    created_at BIGINT,
    updated_at BIGINT,

    CONSTRAINT uq_api_key UNIQUE (tenant_id, team_id)
);