from .spacy_lib import nlp
from .config import PRONOUN_CONTEXT


def raw_input(text: str) -> str:
    print("─" * 50)
    print("STEP 1 — Raw Text Input")
    print(f"  Input: '{text}'")
    return text


def preprocess(text: str) -> str:
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


def resolve_coreferences(text: str, context: str = "") -> str:
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
