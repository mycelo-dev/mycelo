CREATE TABLE DESTINATIONS (
    destination_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    destination_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    tenant_id BIGINT REFERENCES tenants(tenant_id),
    team_id BIGINT REFERENCES teams(team_id), 
    destination_name VARCHAR(255),
    destination_address VARCHAR(500),
    created_at BIGINT,
    updated_at BIGINT,
    
    CONSTRAINT uq_destination_name UNIQUE(tenant_id, team_id, destination_name),
    CONSTRAINT uq_destination_address UNIQUE(tenant_id, team_id, destination_address)
)
;

CREATE TABLE DESTINATION_TOPIC_MAPPING (
    destination_public_id UUID REFERENCES DESTINATIONS(destination_public_id),
    topic_public_id UUID REFERENCES TOPICS(topic_public_id)
)
;