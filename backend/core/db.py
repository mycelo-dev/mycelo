from contextlib import contextmanager

from gqlalchemy import Memgraph
from sqlalchemy import create_engine
from sqlalchemy.orm import Session

from data_to_graph.config import (
    MEMGRAPH_URI, MEMGRAPH_USER, MEMGRAPH_PASSWORD,
    POSTGRES_HOST, POSTGRES_PORT, POSTGRES_DB,
    POSTGRES_USER, POSTGRES_PASSWORD,
)

# ── Memgraph (graph) ─────────────────────────────
_host = MEMGRAPH_URI.replace("bolt://", "").split(":")[0]
_port = int(MEMGRAPH_URI.replace("bolt://", "").split(":")[1]) if ":" in MEMGRAPH_URI.replace("bolt://", "") else 7687

db = Memgraph(host=_host, port=_port, username=MEMGRAPH_USER or "", password=MEMGRAPH_PASSWORD or "")

# ── PostgreSQL (embeddings) ───────────────────────
DATABASE_URL = (
    f"postgresql+psycopg://{POSTGRES_USER}:{POSTGRES_PASSWORD}"
    f"@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"
)

engine = create_engine(DATABASE_URL)


@contextmanager
def get_pg_session():
    session = Session(engine)
    try:
        yield session
        session.commit()
    except Exception:
        session.rollback()
        raise
    finally:
        session.close()
