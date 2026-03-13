from .ner import extract_entities_and_relationships
from .graph import update_graph


def ingest(text: str):
    
    # extract entities and relationsips
    entities, relationships = extract_entities_and_relationships(text)

    # insert those into postgres
    update_graph(entities, relationships)

    pass
    