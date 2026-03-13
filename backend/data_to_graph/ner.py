from .spacy_lib import spacy, nlp

def _get_entities(text: str):

    doc = nlp(text)
    entities = doc.ents
    return entities

def extract_entities_and_relationships(text: str):

    entities = _get_entities(text)
    print(entities)
    relationships = []
    return entities, relationships