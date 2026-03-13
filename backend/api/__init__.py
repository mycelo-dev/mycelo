from fastapi import APIRouter

from .ingest_route import router as ingest_router

router = APIRouter(prefix="/v1")

router.include_router(ingest_router)

__all__ = ["router"]