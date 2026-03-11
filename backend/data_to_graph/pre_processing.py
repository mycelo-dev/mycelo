# ─────────────────────────────────────────────
# STEP 2 — PRE-PROCESSING
# ─────────────────────────────────────────────

def step2_preprocess(text: str) -> str:
    print("\nSTEP 2 — Pre-processing")
    cleaned = " ".join(text.split())
    cleaned = cleaned[0].upper() + cleaned[1:] if cleaned else cleaned

    doc    = nlp(cleaned)
    tokens = []
    for token in doc:
        if token.pos_ == "PROPN" and not token.text[0].isupper():
            tokens.append(token.text.capitalize())
        else:
            tokens.append(token.text)
    cleaned = " ".join(tokens)

    print(f"  Original : '{text}'")
    print(f"  Cleaned  : '{cleaned}'")
    return cleaned