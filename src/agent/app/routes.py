"""API routes for the agent service."""

from fastapi import APIRouter
from fastapi.responses import StreamingResponse

from app.models.requests import (
    AnalyzeRequest,
    CoinSearchRequest,
    CoinShowSearchRequest,
    PortfolioReviewRequest,
)
from app.models.responses import AgentResponse
from app.streaming import stream_graph_events
from app.supervisor import create_supervisor

router = APIRouter(prefix="/api")


def _build_messages(message: str, history: list | None = None, system_prompt: str = ""):
    """Convert request data into LangChain messages."""
    from langchain_core.messages import AIMessage, HumanMessage, SystemMessage

    messages = []
    if system_prompt:
        messages.append(SystemMessage(content=system_prompt))
    for msg in (history or []):
        if msg.role == "user":
            messages.append(HumanMessage(content=msg.content))
        elif msg.role == "assistant":
            messages.append(AIMessage(content=msg.content))
    messages.append(HumanMessage(content=message))
    return messages


@router.post("/search/coins")
async def search_coins(request: CoinSearchRequest):
    """Search for coins using multi-agent pipeline with verification. Streams SSE."""
    messages = _build_messages(request.message, request.history, request.agent_prompt)
    graph = create_supervisor(
        request.llm,
        user_message=request.message,
        agent_prompt=request.agent_prompt,
    )

    async def event_stream():
        async for chunk in stream_graph_events(graph, {"messages": messages}):
            yield chunk

    return StreamingResponse(event_stream(), media_type="text/event-stream")


@router.post("/search/shows")
async def search_shows(request: CoinShowSearchRequest):
    """Search for upcoming coin shows with date verification. Streams SSE."""
    messages = _build_messages(request.message, request.history, request.agent_prompt)
    graph = create_supervisor(
        request.llm,
        user_message=request.message,
        agent_prompt=request.agent_prompt,
    )

    async def event_stream():
        async for chunk in stream_graph_events(graph, {"messages": messages}):
            yield chunk

    return StreamingResponse(event_stream(), media_type="text/event-stream")


@router.post("/analyze", response_model=AgentResponse)
async def analyze_coin(request: AnalyzeRequest):
    """Analyze coin images using vision model."""
    # TODO: Phase 5 — wire to Team 3 (non-streaming, returns structured response)
    return AgentResponse(message="Analysis not yet implemented")


@router.post("/portfolio/review")
async def review_portfolio(request: PortfolioReviewRequest):
    """Review portfolio with live valuation. Streams SSE."""
    prompt = request.valuation_prompt or "You are a numismatic portfolio analyst."
    messages = _build_messages(request.message or "Analyze my portfolio", request.history, prompt)
    graph = create_supervisor(
        request.llm,
        user_message=request.message or "Analyze my portfolio",
        agent_prompt=prompt,
    )

    async def event_stream():
        async for chunk in stream_graph_events(graph, {"messages": messages}):
            yield chunk

    return StreamingResponse(event_stream(), media_type="text/event-stream")

