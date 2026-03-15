from fastapi import APIRouter, BackgroundTasks
from model import IngestRequest

from data_to_graph.ingest import ingest

router = APIRouter(tags=["Ingest"])

@router.post("/ingest")
async def ingest_route(req: IngestRequest, background_tasks: BackgroundTasks):
    background_tasks.add_task(ingest, req.text)
    return {"status": "accepted"}