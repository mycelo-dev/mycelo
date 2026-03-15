from sqlalchemy import text
from sqlalchemy.dialects.postgresql import insert

from core.db import engine, get_pg_session
from core.models import Base, Embedding


def upsert_embedding(
    *, node_id: str, name: str, label: str,
    tenant_id: str, embedding: list[float],
) -> None:
    """Upsert a node embedding in PostgreSQL."""
    with get_pg_session() as pg:
        stmt = insert(Embedding).values(
            node_id=node_id,
            name=name,
            label=label,
            tenant_id=tenant_id,
            embedding=embedding,
        ).on_conflict_do_update(
            index_elements=["name", "label", "tenant_id"],
            set_={"embedding": embedding, "node_id": node_id},
        )
        pg.execute(stmt)


def ensure_pg_schema() -> None:
    """Create pgvector extension and embeddings table."""
    with engine.connect() as conn:
        conn.execute(text("CREATE EXTENSION IF NOT EXISTS vector"))
        conn.commit()

    Base.metadata.create_all(engine)
    print("  Postgres embeddings table ready.")
