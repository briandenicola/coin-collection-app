"""FastAPI entry point for the Ancient Coins multi-agent service."""

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, Query
from fastapi.middleware.cors import CORSMiddleware

from app.config import settings
from app.logging_config import ring_handler, setup_logging
from app.routes import router

# Configure logging before anything else
setup_logging(settings.log_level)

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(application: FastAPI):
    logger.info("Agent service starting (log_level=%s)", settings.log_level)
    yield


app = FastAPI(
    title="Ancient Coins Agent Service",
    version="0.1.0",
    docs_url="/docs" if settings.debug else None,
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router)


@app.get("/health")
async def health():
    return {"status": "ok", "service": "agent"}


@app.get("/logs")
async def get_logs(
    limit: int = Query(default=500, ge=1, le=2000),
    level: str = Query(default=""),
):
    """Return recent log entries from the in-memory ring buffer."""
    logs = ring_handler.get_logs(limit=limit, level=level)
    return {"logs": logs, "count": len(logs), "logLevel": settings.log_level}
