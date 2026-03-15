from datetime import datetime
from typing import Optional

from .models import PipelineResult
from .preprocess import raw_input, preprocess, resolve_coreferences
from .ner import extract_entities_and_relationships
from .entities import resolve_temporal, deduplicate, embed_entities
from .graph import update_graph


def _print_summary(result: PipelineResult) -> None:
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


def ingest(
    text: str,
    context: str = "",
    reference_date: Optional[datetime] = None,
) -> PipelineResult:
    ingested_at = datetime.now().isoformat()

    raw           = raw_input(text)
    cleaned       = preprocess(raw)
    resolved      = resolve_coreferences(cleaned, context)
    entities, relationships = extract_entities_and_relationships(resolved)
    entities      = resolve_temporal(entities, reference_date)
    entities      = deduplicate(entities)
    entities      = embed_entities(entities)
    update_graph(entities, relationships, ingested_at)

    result = PipelineResult(
        original_text=raw,
        cleaned_text=resolved,
        entities=entities,
        relationships=relationships,
        ingested_at=ingested_at,
    )

    _print_summary(result)
    return result
