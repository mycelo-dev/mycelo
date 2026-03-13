from fastapi import FastAPI
from api import router

app = FastAPI(title="mycelo")

app.include_router(router)

