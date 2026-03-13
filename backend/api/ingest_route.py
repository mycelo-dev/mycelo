from fastapi import APIRouter
from model import IngestRequest

from data_to_graph.ingest import ingest as run_ingest

router = APIRouter(tags=["Ingest"])

@router.post("/ingest")
def ingest(req: IngestRequest):
    run_ingest(req.text)
    return {"status": "ok"}