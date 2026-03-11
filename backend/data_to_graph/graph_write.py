# ─────────────────────────────────────────────
# STEP 9 — GRAPH WRITE (Postgres + pgvector)
#
# Schema:
#   nodes(id, name, label, tenant_id, embedding,
#         resolved_date, created_at, updated_at)
#   edges(id, head_id, tail_id, relation, confidence,
#         raw_text, tenant_id, ingested_at)
#
# Semantic search query:
#   SELECT name, label,
#          1 - (embedding <=> %s::vector) AS similarity
#   FROM nodes
#   WHERE tenant_id = 'demo'
#   ORDER BY embedding <=> %s::vector
#   LIMIT 10;
#
# Multi-hop traversal query:
#   WITH RECURSIVE traverse AS (
#     SELECT id, name, 0 AS depth
#     FROM nodes WHERE name = 'Auth service'
#     UNION ALL
#     SELECT n.id, n.name, t.depth + 1
#     FROM nodes n
#     JOIN edges e ON e.head_id = t.id
#     JOIN traverse t ON t.id = e.head_id
#     WHERE t.depth < 5
#   )
#   SELECT * FROM traverse;
# ─────────────────────────────────────────────

def _get_pg_conn():
    conn = psycopg2.connect(
        host=POSTGRES_HOST, port=POSTGRES_PORT,
        dbname=POSTGRES_DB, user=POSTGRES_USER,
        password=POSTGRES_PASSWORD,
    )
    register_vector(conn)
    return conn


def step9_ensure_schema() -> None:
    """Create tables if they don't exist. Call once before first ingest."""
    conn = _get_pg_conn()
    cur  = conn.cursor()
    cur.execute(f"""
        CREATE EXTENSION IF NOT EXISTS vector;

        CREATE TABLE IF NOT EXISTS nodes (
            id            TEXT PRIMARY KEY,
            name          TEXT NOT NULL,
            label         TEXT NOT NULL,
            tenant_id     TEXT NOT NULL,
            resolved_date TEXT,
            embedding     vector({EMBEDDING_DIM}),
            created_at    TEXT,
            updated_at    TEXT,
            UNIQUE (name, label, tenant_id)
        );

        CREATE TABLE IF NOT EXISTS edges (
            id          TEXT PRIMARY KEY,
            head_id     TEXT NOT NULL REFERENCES nodes(id),
            tail_id     TEXT NOT NULL REFERENCES nodes(id),
            relation    TEXT NOT NULL,
            confidence  REAL DEFAULT 1.0,
            raw_text    TEXT,
            tenant_id   TEXT NOT NULL,
            ingested_at TEXT,
            UNIQUE (head_id, tail_id, relation, tenant_id)
        );

        CREATE INDEX IF NOT EXISTS idx_edges_head ON edges(head_id);
        CREATE INDEX IF NOT EXISTS idx_edges_tail ON edges(tail_id);
        CREATE INDEX IF NOT EXISTS idx_nodes_name ON nodes(name, tenant_id);
        CREATE INDEX IF NOT EXISTS idx_nodes_vector
            ON nodes USING ivfflat (embedding vector_cosine_ops)
            WITH (lists = 100);
    """)
    conn.commit()
    cur.close()
    conn.close()
    print("  Postgres schema ready.")


def step9_graph_write(
    entities: list[Entity],
    relationships: list[Relationship],
    ingested_at: str,
    tenant_id: str = TENANT_ID,
    dry_run: bool = True,
) -> None:
    print("\nSTEP 9 — Graph Write (Postgres + pgvector)")

    entity_map = {e.id: (e.resolved_text or e.text) for e in entities}

    if dry_run:
        print("  [DRY RUN — not writing to Postgres]")
        for e in entities:
            name = e.resolved_text or e.text
            has_vec = "✓ embedding" if e.embedding else "no embedding"
            print(f"  Would INSERT node: ({name}) [{e.label}] {has_vec}")
        for r in relationships:
            h = entity_map.get(r.head_id, "?")
            t = entity_map.get(r.tail_id, "?")
            print(f"  Would INSERT edge: ({h}) -[{r.relation}]-> ({t})")
        return

    try:
        conn = _get_pg_conn()
        cur  = conn.cursor()

        # Write nodes + embeddings together
        for entity in entities:
            name      = entity.resolved_text or entity.text
            embedding = np.array(entity.embedding) if entity.embedding else None

            cur.execute("""
                INSERT INTO nodes
                    (id, name, label, tenant_id, resolved_date, embedding, created_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (name, label, tenant_id) DO UPDATE SET
                    updated_at    = EXCLUDED.created_at,
                    embedding     = COALESCE(EXCLUDED.embedding, nodes.embedding),
                    resolved_date = COALESCE(EXCLUDED.resolved_date, nodes.resolved_date)
                RETURNING id
            """, (
                entity.id, name, entity.label, tenant_id,
                entity.resolved_date, embedding, ingested_at,
            ))

            row = cur.fetchone()
            if row:
                entity.id = row[0]   # use existing id if node already existed

            vec_status = "+ embedding" if embedding is not None else ""
            print(f"  Wrote node: ({name}) [{entity.label}] {vec_status}")

        # Rebuild map with updated ids
        entity_map = {e.id: e for e in entities}

        # Write edges
        for rel in relationships:
            head = entity_map.get(rel.head_id)
            tail = entity_map.get(rel.tail_id)
            if not head or not tail:
                continue

            cur.execute("""
                INSERT INTO edges
                    (id, head_id, tail_id, relation, confidence,
                     raw_text, tenant_id, ingested_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (head_id, tail_id, relation, tenant_id) DO UPDATE SET
                    confidence  = EXCLUDED.confidence,
                    ingested_at = EXCLUDED.ingested_at
            """, (
                str(uuid.uuid4()),
                head.id, tail.id, rel.relation,
                rel.confidence, rel.raw_text,
                tenant_id, ingested_at,
            ))

            head_name = head.resolved_text or head.text
            tail_name = tail.resolved_text or tail.text
            print(f"  Wrote edge: ({head_name}) -[{rel.relation}]-> ({tail_name})")

        conn.commit()
        cur.close()
        conn.close()

    except Exception as e:
        print(f"  Postgres error: {e}")