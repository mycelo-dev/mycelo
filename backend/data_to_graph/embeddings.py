# ─────────────────────────────────────────────
# STEP 8 — EMBEDDINGS (sentence-transformers)
# Embeddings are generated here, then written
# to Postgres in step 9 alongside nodes/edges.
# ─────────────────────────────────────────────

def step8_embed(entities: list[Entity]) -> list[Entity]:
    print("\nSTEP 8 — Embeddings (sentence-transformers → stored in Postgres)")

    for entity in entities:
        name   = entity.resolved_text or entity.text
        text   = f"{entity.label}: {name}"
        if entity.resolved_date:
            text += f" ({entity.resolved_date})"
        vector = embedder.encode(text).tolist()
        entity.embedding = vector
        print(
            f"  Embedded '{text}' → "
            f"dim={len(vector)}, "
            f"first3={[round(v, 3) for v in vector[:3]]}"
        )

    return entities