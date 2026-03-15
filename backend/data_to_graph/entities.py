import uuid
from datetime import datetime
from typing import Optional

import dateparser
from gliner import GLiNER
from sentence_transformers import SentenceTransformer

from .models import Entity
from .config import LOCATION_ALIASES

ner_model = GLiNER.from_pretrained("urchade/gliner_large-v2.1")
embedder = SentenceTransformer("all-MiniLM-L6-v2")

NER_LABELS = [
    "person", "location", "date", "time",
    "organization", "software service", "system",
    "product", "concept", "event", "technology",
    "action", "attribute",
]


def extract_entities(text: str) -> list[Entity]:
    print("\nSTEP 4 — Named Entity Recognition (Gliner)")

    raw_entities = ner_model.predict_entities(text, NER_LABELS, threshold=0.4)
    raw_entities = sorted(raw_entities, key=lambda x: -x["score"])

    seen_spans = []
    filtered   = []
    for e in raw_entities:
        span = (e["start"], e["end"])
        if not any(s[0] <= span[0] < s[1] or span[0] <= s[0] < span[1] for s in seen_spans):
            seen_spans.append(span)
            filtered.append(e)

    entities = []
    for e in filtered:
        entity = Entity(
            id=str(uuid.uuid4()),
            text=e["text"],
            label=e["label"],
            confidence=round(e["score"], 3),
        )
        entities.append(entity)
        print(f"  Found: '{entity.text}' → {entity.label} (conf: {entity.confidence})")

    if not entities:
        print("  No entities found.")

    return entities


def resolve_temporal(
    entities: list[Entity],
    reference_date: Optional[datetime] = None,
) -> list[Entity]:
    print("\nSTEP 6 — Temporal Resolution")
    ref = reference_date or datetime.now()

    for entity in entities:
        if entity.label in ("date", "time"):
            parsed = dateparser.parse(
                entity.text,
                settings={
                    "PREFER_DAY_OF_MONTH": "first",
                    "RELATIVE_BASE": ref,
                    "RETURN_AS_TIMEZONE_AWARE": False,
                },
            )
            if parsed:
                entity.resolved_date = parsed.strftime("%Y-%m-%d")
                print(f"  '{entity.text}' → {entity.resolved_date}")
            else:
                entity.resolved_date = entity.text
                print(f"  '{entity.text}' → could not parse, kept as-is")

    return entities


def deduplicate(entities: list[Entity]) -> list[Entity]:
    print("\nSTEP 7 — Deduplication & Alias Resolution")

    for entity in entities:
        lower = entity.text.lower()
        if entity.label == "location" and lower in LOCATION_ALIASES:
            entity.resolved_text = LOCATION_ALIASES[lower]
            print(f"  Alias: '{entity.text}' → '{entity.resolved_text}'")
        else:
            entity.resolved_text = entity.text
        if "UNKNOWN_PERSON" in entity.text:
            entity.resolved_text = "Unknown Person"
            print(f"  Normalized: '{entity.text}' → 'Unknown Person'")

    seen   = {}
    unique = []
    for entity in entities:
        key = (entity.resolved_text.lower(), entity.label)
        if key not in seen:
            seen[key] = entity
            unique.append(entity)
        else:
            print(f"  Duplicate removed: '{entity.text}'")

    return unique


def embed_entities(entities: list[Entity]) -> list[Entity]:
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
