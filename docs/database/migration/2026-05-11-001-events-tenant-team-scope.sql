ALTER TABLE topics
ADD COLUMN IF NOT EXISTS tenant_public_id UUID REFERENCES tenants(tenant_public_id),
ADD COLUMN IF NOT EXISTS team_public_id UUID REFERENCES teams(team_public_id)
;

UPDATE topics topic
SET
    tenant_public_id = tenant.tenant_public_id,
    team_public_id = team.team_public_id
FROM tenants tenant, teams team
WHERE topic.tenant_id = tenant.tenant_id
AND topic.team_id = team.team_id
AND (topic.tenant_public_id IS NULL OR topic.team_public_id IS NULL)
;

ALTER TABLE destinations
ADD COLUMN IF NOT EXISTS tenant_public_id UUID REFERENCES tenants(tenant_public_id),
ADD COLUMN IF NOT EXISTS team_public_id UUID REFERENCES teams(team_public_id)
;

UPDATE destinations destination
SET
    tenant_public_id = tenant.tenant_public_id,
    team_public_id = team.team_public_id
FROM tenants tenant, teams team
WHERE destination.tenant_id = tenant.tenant_id
AND destination.team_id = team.team_id
AND (destination.tenant_public_id IS NULL OR destination.team_public_id IS NULL)
;

ALTER TABLE events
ADD COLUMN IF NOT EXISTS tenant_public_id UUID REFERENCES tenants(tenant_public_id),
ADD COLUMN IF NOT EXISTS team_public_id UUID REFERENCES teams(team_public_id)
;

ALTER TABLE topics
DROP CONSTRAINT IF EXISTS unique_topic_name
;

ALTER TABLE topics
ADD CONSTRAINT unique_topic_name UNIQUE(tenant_public_id, team_public_id, topic_name)
;

ALTER TABLE destinations
DROP CONSTRAINT IF EXISTS uq_destination_name
;

ALTER TABLE destinations
ADD CONSTRAINT uq_destination_name UNIQUE(tenant_public_id, team_public_id, destination_name)
;

ALTER TABLE destinations
DROP CONSTRAINT IF EXISTS uq_destination_address
;

ALTER TABLE destinations
ADD CONSTRAINT uq_destination_address UNIQUE(tenant_public_id, team_public_id, destination_address)
;

CREATE INDEX IF NOT EXISTS idx_topics_tenant_team_public
ON topics (tenant_public_id, team_public_id)
;

CREATE INDEX IF NOT EXISTS idx_destinations_tenant_team_public
ON destinations (tenant_public_id, team_public_id)
;

CREATE INDEX IF NOT EXISTS idx_events_tenant_team_public_topic_created_id
ON events (tenant_public_id, team_public_id, topic, created_at, id)
;
