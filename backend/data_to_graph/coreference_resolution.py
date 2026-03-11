# ─────────────────────────────────────────────
# STEP 3 — CO-REFERENCE RESOLUTION
# ─────────────────────────────────────────────

def step3_coref(text: str, context: str = "") -> str:
    print("\nSTEP 3 — Co-reference Resolution")

    pronouns = {"he", "she", "they", "it", "him", "her"}
    tokens   = text.split()
    resolved_tokens = []
    changed  = False

    context_person = None
    if context:
        doc     = nlp(context)
        persons = [ent.text for ent in doc.ents if ent.label_ == "PERSON"]
        if persons:
            context_person = persons[-1]

    for token in tokens:
        clean = token.lower().strip(".,!?")
        if clean in pronouns:
            if clean in PRONOUN_CONTEXT:
                resolved_tokens.append(PRONOUN_CONTEXT[clean])
                print(f"  Resolved: '{token}' → '{PRONOUN_CONTEXT[clean]}' (context map)")
                changed = True
            elif context_person:
                resolved_tokens.append(context_person)
                print(f"  Resolved: '{token}' → '{context_person}' (prior sentence)")
                changed = True
            else:
                resolved_tokens.append("UNKNOWN_PERSON")
                print(f"  No context for '{token}' → UNKNOWN_PERSON")
                changed = True
        else:
            resolved_tokens.append(token)

    resolved = " ".join(resolved_tokens)
    if not changed:
        print("  No pronouns found — text unchanged")
    print(f"  Output: '{resolved}'")
    return resolved