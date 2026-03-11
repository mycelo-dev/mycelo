# ─────────────────────────────────────────────
# STEP 10 — FINAL GRAPH SUMMARY
# ─────────────────────────────────────────────

def step10_summary(result: PipelineResult) -> None:
    print("\nSTEP 10 — Final Graph Summary")
    print("─" * 50)
    print(f"  Original : '{result.original_text}'")
    print(f"  Resolved : '{result.cleaned_text}'")
    print(f"  Ingested : {result.ingested_at}")

    print(f"\n  Nodes ({len(result.entities)}):")
    for e in result.entities:
        name     = e.resolved_text or e.text
        extra    = f" → {e.resolved_date}" if e.resolved_date else ""
        has_vec  = " [embedded]" if e.embedding else ""
        print(f"    [{e.label.upper()}] {name}{extra}{has_vec}")

    entity_map = {e.id: (e.resolved_text or e.text) for e in result.entities}

    print(f"\n  Edges ({len(result.relationships)}):")
    for r in result.relationships:
        h = entity_map.get(r.head_id, "?")
        t = entity_map.get(r.tail_id, "?")
        print(f"    ({h}) ─[{r.relation}]→ ({t})")

    print("\n  Graph:")
    for r in result.relationships:
        h = entity_map.get(r.head_id, "?")
        t = entity_map.get(r.tail_id, "?")
        print(f"    {h}")
        print(f"      └─[{r.relation}]──► {t}")
    print()