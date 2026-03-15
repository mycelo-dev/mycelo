import uuid
from typing import Optional

from .models import Entity, Relationship
from .spacy_lib import nlp
from .rebel_lib import re_model, parse_rebel_output


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


def extract_relations(text: str, entities: list[Entity]) -> list[Relationship]:
    print("\nSTEP 5 — Relation Extraction (REBEL + spaCy fallback)")

    entity_map    = {e.text.lower(): e for e in entities}
    relationships = []

    try:
        output     = re_model(text)
        raw_output = output[0]["generated_text"] if output else ""
        triplets   = parse_rebel_output(raw_output)

        for t in triplets:
            head_e = entity_map.get(t["head"].lower())
            tail_e = entity_map.get(t["tail"].lower())

            if not head_e:
                head_e = Entity(id=str(uuid.uuid4()), text=t["head"], label="unknown")
                entities.append(head_e)
                entity_map[t["head"].lower()] = head_e
            if not tail_e:
                tail_e = Entity(id=str(uuid.uuid4()), text=t["tail"], label="unknown")
                entities.append(tail_e)
                entity_map[t["tail"].lower()] = tail_e

            rel = Relationship(
                head_id=head_e.id,
                relation=t["relation"].upper().replace(" ", "_"),
                tail_id=tail_e.id,
                raw_text=text,
            )
            relationships.append(rel)
            print(f"  REBEL: '{head_e.text}' ─[{rel.relation}]→ '{tail_e.text}'")

    except Exception as e:
        print(f"  REBEL failed ({e}) — spaCy only")

    if len(entities) >= 2:
        spacy_rels  = _infer_from_spacy(text, entities)
        existing    = {(r.head_id, r.tail_id) for r in relationships}
        for r in spacy_rels:
            if (r.head_id, r.tail_id) not in existing:
                relationships.append(r)

    if not relationships:
        print("  No relations found.")

    return relationships
