"""API routes for the agent service."""

from fastapi import APIRouter

from app.models.requests import (
    AnalyzeRequest,
    CoinSearchRequest,
    CoinShowSearchRequest,
    PortfolioReviewRequest,
)
from app.models.responses import AgentResponse

router = APIRouter(prefix="/api")


@router.post("/search/coins", response_model=AgentResponse)
async def search_coins(request: CoinSearchRequest):
    """Search for coins using multi-agent pipeline with verification."""
    # TODO: Phase 3 — wire to Team 1 supervisor
    return AgentResponse(message="Coin search not yet implemented", suggestions=[])


@router.post("/search/shows", response_model=AgentResponse)
async def search_shows(request: CoinShowSearchRequest):
    """Search for upcoming coin shows with date verification."""
    # TODO: Phase 4 — wire to Team 2 supervisor
    return AgentResponse(message="Coin show search not yet implemented", shows=[])


@router.post("/analyze", response_model=AgentResponse)
async def analyze_coin(request: AnalyzeRequest):
    """Analyze coin images using vision model."""
    # TODO: Phase 5 — wire to Team 3 supervisor
    return AgentResponse(message="Analysis not yet implemented")


@router.post("/portfolio/review", response_model=AgentResponse)
async def review_portfolio(request: PortfolioReviewRequest):
    """Review portfolio with live valuation."""
    # TODO: Phase 6 — wire to Team 4 supervisor
    return AgentResponse(message="Portfolio review not yet implemented")
