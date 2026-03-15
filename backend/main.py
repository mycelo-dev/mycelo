from fastapi import FastAPI
from api import router
from repository import ensure_graph_schema, ensure_pg_schema

app = FastAPI(title="mycelo")

app.include_router(router)


@app.on_event("startup")
def startup():
    ensure_graph_schema()
    ensure_pg_schema()
