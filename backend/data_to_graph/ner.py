from .models import Entity, Relationship
from .entities import extract_entities
from .relations import extract_relations


def extract_entities_and_relationships(text: str) -> tuple[list[Entity], list[Relationship]]:
    entities = extract_entities(text)
    relationships = extract_relations(text, entities)
    return entities, relationships
