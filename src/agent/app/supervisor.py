"""Top-level supervisor that routes requests to the appropriate team.

The supervisor examines the user's message and delegates to:
- Team 1 (Coin Search) for finding coins to buy
- Team 2 (Coin Shows) for finding upcoming shows/events
- Team 3 (Coin Analysis) for analyzing coin images
- Team 4 (Portfolio Review) for portfolio analysis and valuation
"""

from typing import Literal

from langchain_core.messages import HumanMessage, SystemMessage
from langgraph.graph import END, MessagesState, StateGraph
from langgraph.types import Command

from app.llm.provider import get_chat_model
from app.models.requests import LLMConfig, UserContext
from app.teams.coin_search import create_coin_search_team
from app.teams.coin_shows import create_coin_show_team

ROUTER_PROMPT = """You are a routing agent for a numismatic (coin collecting) application.
Your ONLY job is to classify the user's request into exactly one category.

Respond with ONLY one of these words:
- "coin_search" — if the user wants to find, buy, or search for coins
- "coin_shows" — if the user asks about coin shows, conventions, expos, or events
- "analysis" — if the user wants to analyze coin images or get AI analysis of a coin
- "portfolio" — if the user wants portfolio analysis, collection review, or valuation
- "general" — if the request doesn't fit the above categories

Respond with ONLY the category word, nothing else."""


def create_router(llm_config: LLMConfig):
    """Create a lightweight router that classifies intent."""
    model = get_chat_model(llm_config)

    RouteTarget = Literal["coin_search", "coin_shows", "analysis", "portfolio", "general"]

    async def route_request(state: MessagesState) -> Command[RouteTarget]:
        messages = [
            SystemMessage(content=ROUTER_PROMPT),
            HumanMessage(content=state["messages"][-1].content if state["messages"] else ""),
        ]
        response = await model.ainvoke(messages)
        route = response.content.strip().lower().replace('"', "").replace("'", "")

        valid_routes = {"coin_search", "coin_shows", "analysis", "portfolio", "general"}
        if route not in valid_routes:
            route = "general"

        return Command(goto=route)

    return route_request


def create_supervisor(
    llm_config: LLMConfig,
    user_message: str = "",
    agent_prompt: str = "",
    user_context: UserContext | None = None,
    analysis_node=None,
    portfolio_node=None,
):
    """Build the top-level supervisor graph.

    Team 1 (coin_search) and Team 2 (coin_shows) are always wired.
    Other teams are optional — passthrough is used if not provided.
    """

    # Build Team 1 as a callable node
    coin_search_graph = create_coin_search_team(llm_config, user_prompt=agent_prompt)

    async def coin_search_node(state: MessagesState) -> dict:
        """Delegate to Team 1 coin search pipeline."""
        result = await coin_search_graph.ainvoke({
            "messages": [],
            "search_results": "",
            "verification_results": "",
            "formatted_results": "",
            "user_message": user_message,
        })
        return {"messages": result.get("messages", [])}

    # Build Team 2 as a callable node
    coin_show_graph = create_coin_show_team(
        llm_config, user_context=user_context, agent_prompt=agent_prompt,
    )

    async def coin_shows_node(state: MessagesState) -> dict:
        """Delegate to Team 2 coin show search pipeline."""
        result = await coin_show_graph.ainvoke({
            "messages": [],
            "search_results": "",
            "verification_results": "",
            "formatted_results": "",
            "user_message": user_message,
            "location_context": "",
        })
        return {"messages": result.get("messages", [])}

    async def passthrough(state: MessagesState) -> dict:
        """Placeholder for teams not yet implemented."""
        from langchain_core.messages import AIMessage

        return {"messages": [AIMessage(content="This capability is not yet available. Please try again later.")]}

    async def general_handler(state: MessagesState) -> dict:
        """Handle general questions using the base LLM."""
        model = get_chat_model(llm_config)
        response = await model.ainvoke(state["messages"])
        return {"messages": [response]}

    router = create_router(llm_config)

    graph = StateGraph(MessagesState)

    graph.add_node("router", router)
    graph.add_node("coin_search", coin_search_node)
    graph.add_node("coin_shows", coin_shows_node)
    graph.add_node("analysis", analysis_node or passthrough)
    graph.add_node("portfolio", portfolio_node or passthrough)
    graph.add_node("general", general_handler)

    graph.set_entry_point("router")

    graph.add_edge("coin_search", END)
    graph.add_edge("coin_shows", END)
    graph.add_edge("analysis", END)
    graph.add_edge("portfolio", END)
    graph.add_edge("general", END)

    return graph.compile()
