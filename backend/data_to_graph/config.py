import os

from dotenv import load_dotenv

load_dotenv()

# Memgraph (graph)
MEMGRAPH_URI      = os.getenv("MEMGRAPH_URI", "bolt://localhost:7687")
MEMGRAPH_USER     = os.getenv("MEMGRAPH_USER", "")
MEMGRAPH_PASSWORD = os.getenv("MEMGRAPH_PASSWORD", "")

# PostgreSQL (embeddings)
POSTGRES_HOST     = os.getenv("POSTGRES_HOST", "localhost")
POSTGRES_PORT     = int(os.getenv("POSTGRES_PORT", "5432"))
POSTGRES_DB       = os.getenv("POSTGRES_DB", "mycelo")
POSTGRES_USER     = os.getenv("POSTGRES_USER", "postgres")
POSTGRES_PASSWORD = os.getenv("POSTGRES_PASSWORD", "password")

TENANT_ID         = "demo"
EMBEDDING_DIM     = 384   # all-MiniLM-L6-v2 output size

LOCATION_ALIASES = {
    "bombay":   "Mumbai",
    "calcutta": "Kolkata",
    "madras":   "Chennai",
    "peking":   "Beijing",
}

PRONOUN_CONTEXT = {}   # e.g. {"he": "Rahul"}
