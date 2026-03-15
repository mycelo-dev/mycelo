import numpy as np

from repository.graph_repository import upsert_node, upsert_edge
from repository.postgres_repository import upsert_embedding
from .models import Entity, Relationship
from .config import TENANT_ID


def update_graph(
    entities: list[Entity],
    relationships: list[Relationship],
    ingested_at: str,
    tenant_id: str = TENANT_ID,
) -> None:
    print("\nSTEP 9 — Graph Write (Memgraph + Postgres embeddings)")

    try:
        # ── Nodes → Memgraph ─────────────────────
        for entity in entities:
            name = entity.resolved_text or entity.text

            returned_id = upsert_node(
                node_id=entity.id,
                name=name,
                label=entity.label,
                tenant_id=tenant_id,
                resolved_date=entity.resolved_date,
                created_at=ingested_at,
            )
            if returned_id:
                entity.id = returned_id

            print(f"  Wrote node: ({name}) [{entity.label}]")

        # ── Embeddings → PostgreSQL ───────────────
        for entity in entities:
            if not entity.embedding:
                continue
            name = entity.resolved_text or entity.text
            embedding = np.array(entity.embedding).tolist()

            upsert_embedding(
                node_id=entity.id,
                name=name,
                label=entity.label,
                tenant_id=tenant_id,
                embedding=embedding,
            )
            print(f"  Wrote embedding: ({name}) [{entity.label}]")

        # ── Edges → Memgraph ─────────────────────
        entity_map = {e.id: e for e in entities}

        for rel in relationships:
            head = entity_map.get(rel.head_id)
            tail = entity_map.get(rel.tail_id)
            if not head or not tail:
                continue

            head_name = head.resolved_text or head.text
            tail_name = tail.resolved_text or tail.text

            upsert_edge(
                head_name=head_name,
                head_label=head.label,
                tail_name=tail_name,
                tail_label=tail.label,
                tenant_id=tenant_id,
                relation=rel.relation,
                edge_id=rel.head_id + "_" + rel.tail_id + "_" + rel.relation,
                confidence=rel.confidence,
                raw_text=rel.raw_text,
                ingested_at=ingested_at,
            )

            print(f"  Wrote edge: ({head_name}) -[{rel.relation}]-> ({tail_name})")

    except Exception as e:
        print(f"  Graph write error: {e}")
