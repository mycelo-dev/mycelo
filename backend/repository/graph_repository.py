from core.db import db


_UPSERT_NODE = """
MERGE (n:Node {name: $name, label: $label, tenant_id: $tenant_id})
ON CREATE SET
    n.id            = $id,
    n.resolved_date = $resolved_date,
    n.created_at    = $created_at,
    n.updated_at    = $created_at
ON MATCH SET
    n.updated_at    = $created_at,
    n.resolved_date = $resolved_date
RETURN n.id AS id
"""

_UPSERT_EDGE = """
MATCH (h:Node {name: $head_name, label: $head_label, tenant_id: $tenant_id})
MATCH (t:Node {name: $tail_name, label: $tail_label, tenant_id: $tenant_id})
MERGE (h)-[r:RELATES {relation: $relation, tenant_id: $tenant_id}]->(t)
ON CREATE SET
    r.id          = $id,
    r.confidence  = $confidence,
    r.raw_text    = $raw_text,
    r.ingested_at = $ingested_at
ON MATCH SET
    r.confidence  = $confidence,
    r.ingested_at = $ingested_at
"""


def upsert_node(
    *, node_id: str, name: str, label: str, tenant_id: str,
    resolved_date: str | None, created_at: str,
) -> str | None:
    """Upsert a node in Memgraph. Returns the node id."""
    results = db.execute_and_fetch(
        _UPSERT_NODE,
        {
            "name": name,
            "label": label,
            "tenant_id": tenant_id,
            "id": node_id,
            "resolved_date": resolved_date,
            "created_at": created_at,
        },
    )
    for record in results:
        return record["id"]
    return None


def upsert_edge(
    *, head_name: str, head_label: str, tail_name: str, tail_label: str,
    tenant_id: str, relation: str, edge_id: str,
    confidence: float, raw_text: str, ingested_at: str,
) -> None:
    """Upsert an edge in Memgraph."""
    db.execute(
        _UPSERT_EDGE,
        {
            "head_name": head_name,
            "head_label": head_label,
            "tail_name": tail_name,
            "tail_label": tail_label,
            "tenant_id": tenant_id,
            "relation": relation,
            "id": edge_id,
            "confidence": confidence,
            "raw_text": raw_text,
            "ingested_at": ingested_at,
        },
    )


def ensure_graph_schema() -> None:
    """Create indexes in Memgraph."""
    index_queries = [
        "CREATE INDEX ON :Node(name)",
        "CREATE INDEX ON :Node(tenant_id)",
        "CREATE INDEX ON :Node(label)",
        "CREATE INDEX ON :RELATES(tenant_id)",
    ]
    for q in index_queries:
        try:
            db.execute(q)
        except Exception:
            pass  # index already exists
    print("  Memgraph schema ready.")
