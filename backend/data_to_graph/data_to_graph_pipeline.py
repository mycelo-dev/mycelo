# ─────────────────────────────────────────────
# FULL PIPELINE
# ─────────────────────────────────────────────

def run_pipeline(
    text: str,
    context: str = "",
    reference_date: Optional[datetime] = None,
    dry_run: bool = True,
) -> PipelineResult:
    ingested_at = datetime.now().isoformat()

    raw           = step1_raw_input(text)
    cleaned       = step2_preprocess(raw)
    resolved      = step3_coref(cleaned, context)
    entities      = step4_ner(resolved)
    relationships = step5_relation_extraction(resolved, entities)
    entities      = step6_temporal_resolution(entities, reference_date)
    entities      = step7_deduplication(entities)
    entities      = step8_embed(entities)              # embeddings first
    step9_graph_write(entities, relationships,          # then write everything together
                      ingested_at, dry_run=dry_run)

    result = PipelineResult(
        original_text=raw,
        cleaned_text=resolved,
        entities=entities,
        relationships=relationships,
        ingested_at=ingested_at,
    )

    step10_summary(result)
    return result