CREATE TABLE TOPICS (
    topic_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    topic_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    tenant_id BIGINT,
    team_id BIGINT,
    topic_name VARCHAR(255),
    created_at BIGINT,
    updated_at BIGINT
)
;

ALTER TABLE TOPICS ADD CONSTRAINT unique_topic_name UNIQUE(tenant_id, team_id, topic_name)
;

CREATE TABLE TENANTS(
    tenant_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    tenant_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    tenant_name VARCHAR(255),
    created_at BIGINT,
    updated_at BIGINT
)
;

CREATE TABLE TEAMS(
    team_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    team_public_id UUID DEFAULT GEN_RANDOM_UUID() UNIQUE,
    team_name VARCHAR(255),
    created_at BIGINT,
    updated_at BIGINT,
    tenant_id BIGINT REFERENCES TENANTS(tenant_id)
)
;