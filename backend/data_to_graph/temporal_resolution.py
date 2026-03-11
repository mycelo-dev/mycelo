# ─────────────────────────────────────────────
# STEP 6 — TEMPORAL RESOLUTION
# ─────────────────────────────────────────────

def step6_temporal_resolution(
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