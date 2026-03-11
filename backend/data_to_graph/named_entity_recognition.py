# ─────────────────────────────────────────────
# STEP 4 — NAMED ENTITY RECOGNITION (Gliner)
# ─────────────────────────────────────────────

def step4_ner(text: str) -> list[Entity]:
    print("\nSTEP 4 — Named Entity Recognition (Gliner)")

    labels = [
        "person", "location", "date", "time",
        "organization", "software service", "system",
        "product", "concept", "event", "technology",
        "action", "attribute",
    ]

    raw_entities = ner_model.predict_entities(text, labels, threshold=0.4)
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