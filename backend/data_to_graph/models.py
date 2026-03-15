from dataclasses import dataclass, field
from typing import Optional


@dataclass
class Entity:
    id: str
    text: str
    label: str
    resolved_text: Optional[str] = None
    resolved_date: Optional[str] = None
    confidence: float = 1.0
    embedding: list = field(default_factory=list)


@dataclass
class Relationship:
    head_id: str
    relation: str
    tail_id: str
    confidence: float = 1.0
    raw_text: str = ""


@dataclass
class PipelineResult:
    original_text: str
    cleaned_text: str
    entities: list[Entity]
    relationships: list[Relationship]
    ingested_at: str
