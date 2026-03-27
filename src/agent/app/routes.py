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
from app.teams.coin_analysis import create_coin_analysis_team

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
    messages = _build_messages(request.message, request.history)
    graph = create_supervisor(
        request.llm,
        user_message=request.message,
        coin_search_prompt=request.coin_search_prompt,
        coin_shows_prompt=request.coin_shows_prompt,
        user_context=request.user,
        portfolio=request.portfolio,
    )

    async def event_stream():
        async for chunk in stream_graph_events(graph, {"messages": messages}):
            yield chunk

    return StreamingResponse(event_stream(), media_type="text/event-stream")


@router.post("/search/shows")
async def search_shows(request: CoinShowSearchRequest):
    """Search for upcoming coin shows with date verification. Streams SSE."""
    messages = _build_messages(request.message, request.history)
    graph = create_supervisor(
        request.llm,
        user_message=request.message,
        coin_search_prompt=request.coin_search_prompt,
        coin_shows_prompt=request.coin_shows_prompt,
        user_context=request.user,
    )

    async def event_stream():
        async for chunk in stream_graph_events(graph, {"messages": messages}):
            yield chunk

    return StreamingResponse(event_stream(), media_type="text/event-stream")


@router.post("/analyze", response_model=AgentResponse)
async def analyze_coin(request: AnalyzeRequest):
    """Analyze coin images using vision model. Returns structured response."""
    graph = create_coin_analysis_team(
        llm_config=request.llm,
        coin=request.coin,
        images=request.images,
        side=request.side,
        custom_prompt=request.prompt,
    )
    # Don't pass image_contents, coin_context, or analysis_prompt in the
    # initial state — the team constructor captures them via closure and
    # the nodes use state.get(key, closure_default). Passing empty values
    # here would override the closure defaults.
    result = await graph.ainvoke({
        "messages": [],
        "raw_analysis": "",
        "formatted_analysis": "",
    })
    analysis_text = result.get("formatted_analysis", "")
    if not analysis_text:
        # Fall back to messages
        msgs = result.get("messages", [])
        if msgs:
            analysis_text = msgs[-1].content if hasattr(msgs[-1], "content") else str(msgs[-1])

    return AgentResponse(analysis=analysis_text)


@router.post("/portfolio/review")
async def review_portfolio(request: PortfolioReviewRequest):
    """Review portfolio with live valuation. Streams SSE."""
    prompt = request.valuation_prompt or "You are a numismatic portfolio analyst."
    messages = _build_messages(request.message or "Analyze my portfolio", request.history, prompt)
    graph = create_supervisor(
        request.llm,
        user_message=request.message or "Analyze my portfolio",
        user_context=request.user,
        portfolio=request.portfolio,
    )

    async def event_stream():
        async for chunk in stream_graph_events(graph, {"messages": messages}):
            yield chunk

    return StreamingResponse(event_stream(), media_type="text/event-stream")

