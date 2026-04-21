CREATE TABLE DESTINATION (
    destination_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    destination_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    tenant_id BIGINT,
    team_id BIGINT,
    destination_name VARCHAR(255),
    destination_address VARCHAR(500),
    created_at BIGINT,
    updated_at BIGINT,
    
    CONSTRAINT uq_destination_name UNIQUE(tenant_id, team_id, destination_name),
    CONSTRAINT uq_destination_address UNIQUE(tenant_id, team_id, destination_address)
)