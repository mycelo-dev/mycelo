UPDATE topics topic
SET
    tenant_id = tenant.tenant_id,
    team_id = team.team_id
FROM tenants tenant
INNER JOIN teams team
    ON team.tenant_id = tenant.tenant_id
WHERE topic.tenant_public_id = tenant.tenant_public_id
AND topic.team_public_id = team.team_public_id
AND (topic.tenant_id IS NULL OR topic.team_id IS NULL);

UPDATE destinations destination
SET
    tenant_id = tenant.tenant_id,
    team_id = team.team_id
FROM tenants tenant
INNER JOIN teams team
    ON team.tenant_id = tenant.tenant_id
WHERE destination.tenant_public_id = tenant.tenant_public_id
AND destination.team_public_id = team.team_public_id
AND (destination.tenant_id IS NULL OR destination.team_id IS NULL);
