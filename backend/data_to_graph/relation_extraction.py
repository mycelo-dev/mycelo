# ─────────────────────────────────────────────
# STEP 5 — RELATION EXTRACTION (REBEL + spaCy)
# ─────────────────────────────────────────────

def _parse_rebel_output(text: str) -> list[dict]:
    """
    REBEL format: <triplet> HEAD <subj> TAIL <obj> RELATION
    """
    triplets = []
    current  = {}
    mode     = None

    for token in text.replace("<s>", "").replace("<pad>", "").split():
        if token == "<triplet>":
            if current and "_buf" in current and mode:
                current[mode] = current["_buf"].strip()
            if current:
                triplets.append(current)
            current = {"_buf": ""}
            mode = "head"
        elif token == "<subj>":
            current["head"] = current.get("_buf", "").strip()
            current["_buf"] = ""
            mode = "tail"
        elif token == "<obj>":
            current["tail"] = current.get("_buf", "").strip()
            current["_buf"] = ""
            mode = "relation"
        else:
            current["_buf"] = current.get("_buf", "") + " " + token

    if current and "_buf" in current and mode:
        current[mode] = current["_buf"].strip()
    if current:
        triplets.append(current)

    return [
        {
            "head":     t["head"],
            "relation": t["relation"].replace("</s>", "").strip(),
            "tail":     t["tail"],
        }
        for t in triplets
        if t.get("head") and t.get("relation") and t.get("tail")
    ]


def _match_entity(token_text: str, entity_map: dict, entities: list[Entity]) -> Optional[Entity]:
    if token_text.lower() in entity_map:
        return entity_map[token_text.lower()]
    for e in entities:
        if token_text.lower() in e.text.lower() or e.text.lower() in token_text.lower():
            return e
    return None


def _infer_from_spacy(text: str, entities: list[Entity]) -> list[Relationship]:
    doc        = nlp(text)
    entity_map = {e.text.lower(): e for e in entities}
    relationships = []

    def add(head_e, tail_e, relation, source):
        if head_e and tail_e and head_e.id != tail_e.id:
            relationships.append(Relationship(
                head_id=head_e.id, relation=relation,
                tail_id=tail_e.id, raw_text=text, confidence=0.7,
            ))
            print(f"  Inferred ({source}): '{head_e.text}' ─[{relation}]→ '{tail_e.text}'")

    for token in doc:
        # Pattern 1 — VERB: "auth service calls stripe"
        if token.pos_ == "VERB":
            subjects = [c for c in token.children if c.dep_ in ("nsubj", "nsubjpass")]
            objects  = [c for c in token.children if c.dep_ in ("dobj", "attr", "agent")]
            for child in token.children:
                if child.dep_ == "prep":
                    for gc in child.children:
                        if gc.dep_ == "pobj":
                            objects.append(gc)
            parts = [token.lemma_.upper()]
            for c in token.children:
                if c.dep_ == "prt":
                    parts.append(c.text.upper())
            relation = "_".join(parts)
            for s in subjects:
                for o in objects:
                    add(
                        _match_entity(s.text, entity_map, entities),
                        _match_entity(o.text, entity_map, entities),
                        relation, "verb"
                    )

        # Pattern 2 — ADJ: "auth service is dependent on stripe"
        if token.pos_ == "ADJ":
            subjects = []
            if token.head.pos_ == "VERB":
                subjects = [c for c in token.head.children if c.dep_ in ("nsubj", "nsubjpass")]
            objects = []
            for child in token.children:
                if child.dep_ == "prep":
                    for gc in child.children:
                        if gc.dep_ == "pobj":
                            objects.append(gc)
            prep = next((c for c in token.children if c.dep_ == "prep"), None)
            relation = f"{token.lemma_.upper()}{'_' + prep.text.upper() if prep else ''}"
            for s in subjects:
                for o in objects:
                    add(
                        _match_entity(s.text, entity_map, entities),
                        _match_entity(o.text, entity_map, entities),
                        relation, "adj"
                    )

        # Pattern 3 — NOUN: "auth service has a dependency on stripe"
        if token.pos_ == "NOUN" and token.dep_ in ("dobj", "attr", "nsubj"):
            subjects = []
            if token.head.pos_ == "VERB":
                subjects = [c for c in token.head.children if c.dep_ in ("nsubj", "nsubjpass")]
            objects = []
            for child in token.children:
                if child.dep_ == "prep":
                    for gc in child.children:
                        if gc.dep_ == "pobj":
                            objects.append(gc)
            for s in subjects:
                for o in objects:
                    add(
                        _match_entity(s.text, entity_map, entities),
                        _match_entity(o.text, entity_map, entities),
                        token.lemma_.upper(), "noun"
                    )

    return relationships