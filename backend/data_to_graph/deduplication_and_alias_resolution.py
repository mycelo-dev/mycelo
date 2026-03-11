# ─────────────────────────────────────────────
# STEP 7 — ENTITY DEDUPLICATION + ALIAS RESOLUTION
# ─────────────────────────────────────────────

def step7_deduplication(entities: list[Entity]) -> list[Entity]:
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