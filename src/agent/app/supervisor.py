"""Top-level supervisor that routes requests to the appropriate team.

The supervisor examines the user's message and delegates to:
- Team 1 (Coin Search) for finding coins to buy
- Team 2 (Coin Shows) for finding upcoming shows/events
- Team 3 (Coin Analysis) for analyzing coin images
- Team 4 (Portfolio Review) for portfolio analysis and valuation
"""

from typing import Literal

from langchain_core.messages import SystemMessage
from langgraph.graph import END, MessagesState, StateGraph
from langgraph.types import Command

from app.llm.provider import get_chat_model
from app.models.requests import LLMConfig, PortfolioSummary, UserContext
from app.teams.coin_search import create_coin_search_team
from app.teams.coin_shows import create_coin_show_team
from app.teams.portfolio_review import create_portfolio_review_team

ROUTER_PROMPT = """You are a routing agent for a numismatic (coin collecting) application.
Your ONLY job is to classify the user's request into exactly one category.

You will receive the conversation history. Use it to understand context — for example,
if the assistant just asked for the user's location to search for coin shows, and the
user replies with a ZIP code or city, that should be routed to "coin_shows" not "general".

Respond with ONLY one of these words:
- "coin_search" — if the user wants to find, buy, or search for coins
- "coin_shows" — if the user asks about coin shows, conventions, expos, or events,
  OR if the user is providing location info following a coin shows conversation
- "analysis" — if the user wants to analyze coin images or get AI analysis of a coin
- "portfolio" — if the user wants portfolio analysis, collection review, or valuation
- "general" — if the request doesn't fit the above categories

Respond with ONLY the category word, nothing else."""


def create_router(llm_config: LLMConfig):
    """Create a lightweight router that classifies intent."""
    model = get_chat_model(llm_config)

    RouteTarget = Literal["coin_search", "coin_shows", "analysis", "portfolio", "general"]

    async def route_request(state: MessagesState) -> Command[RouteTarget]:
        # Include recent history for context (last 4 messages max to keep it light)
        recent = state["messages"][-4:] if len(state["messages"]) > 4 else state["messages"]
        messages = [SystemMessage(content=ROUTER_PROMPT)] + recent
        response = await model.ainvoke(messages)
        content = response.content if isinstance(response.content, str) else str(response.content)
        route = content.strip().lower().replace('"', "").replace("'", "")

        valid_routes = {"coin_search", "coin_shows", "analysis", "portfolio", "general"}
        if route not in valid_routes:
            route = "general"

        return Command(goto=route)

    return route_request


def create_supervisor(
    llm_config: LLMConfig,
    user_message: str = "",
    coin_search_prompt: str = "",
    coin_shows_prompt: str = "",
    user_context: UserContext | None = None,
    portfolio: PortfolioSummary | None = None,
    analysis_node=None,
):
    """Build the top-level supervisor graph.

    Teams 1 (coin_search), 2 (coin_shows), and 4 (portfolio) are always wired.
    Team 3 (analysis) requires images and uses a direct endpoint.
    """

    # Build Team 1 as a callable node
    coin_search_graph = create_coin_search_team(llm_config, search_prompt=coin_search_prompt)

    async def coin_search_node(state: MessagesState) -> dict:
        """Delegate to Team 1 coin search pipeline."""
        result = await coin_search_graph.ainvoke({
            "messages": [],
            "search_results": "",
            "fetched_listings": "",
            "user_message": user_message,
        })
        return {"messages": result.get("messages", [])}

    # Build Team 2 as a callable node
    coin_show_graph = create_coin_show_team(
        llm_config, user_context=user_context, search_prompt=coin_shows_prompt,
    )

    async def coin_shows_node(state: MessagesState) -> dict:
        """Delegate to Team 2 coin show search pipeline.

        If the user has no ZIP code and hasn't provided a location in their
        message, ask them where they'd like to search before running the team.
        """
        import re

        from langchain_core.messages import AIMessage

        has_zip = user_context and user_context.zip_code
        if not has_zip:
            msg_lower = user_message.lower()
            location_keywords = [
                "near ", "in ", "around ", "close to ",
                "zip ", "zipcode", "zip code",
            ]
            has_location_in_msg = any(kw in msg_lower for kw in location_keywords)

            # Also detect bare ZIP codes (5 digits) or city/state patterns
            has_zip_pattern = bool(re.search(r'\b\d{5}\b', user_message))
            has_location_in_msg = has_location_in_msg or has_zip_pattern

            if not has_location_in_msg:
                return {
                    "messages": [
                        AIMessage(
                            content=(
                                "I'd be happy to find upcoming coin shows for you. "
                                "Could you tell me your city, state, or ZIP code so I "
                                "can prioritize shows in your area?\n\n"
                                "You can also set your ZIP code in **Settings** so I'll "
                                "remember it for next time."
                            )
                        )
                    ]
                }

        location_ctx = ""
        if has_zip:
            location_ctx = f"User is near ZIP code {user_context.zip_code}."
        else:
            # Extract location hint from the user's message
            zip_match = re.search(r'\b(\d{5})\b', user_message)
            if zip_match:
                location_ctx = f"User is near ZIP code {zip_match.group(1)}."
            else:
                location_ctx = f"User indicated their location as: {user_message}"

        result = await coin_show_graph.ainvoke({
            "messages": [],
            "search_results": "",
            "verification_results": "",
            "formatted_results": "",
            "user_message": user_message,
            "location_context": location_ctx,
        })
        return {"messages": result.get("messages", [])}

    # Build Team 4 as a callable node
    portfolio_graph = create_portfolio_review_team(
        llm_config, portfolio=portfolio, user_message=user_message,
    )

    async def portfolio_node(state: MessagesState) -> dict:
        """Delegate to Team 4 portfolio review pipeline."""
        result = await portfolio_graph.ainvoke({
            "messages": [],
            "portfolio_summary": "",
            "valuation_commentary": "",
            "final_analysis": "",
            "user_message": user_message,
        })
        return {"messages": result.get("messages", [])}

    async def passthrough(state: MessagesState) -> dict:
        """Placeholder for teams not yet implemented."""
        from langchain_core.messages import AIMessage

        return {"messages": [AIMessage(content="This capability is not yet available. Please try again later.")]}

    async def general_handler(state: MessagesState) -> dict:
        """Handle general questions with awareness of app capabilities."""
        general_model = get_chat_model(llm_config)
        general_system = (
            "You are a knowledgeable numismatist assistant in a coin collecting "
            "application. You are enthusiastic but informative, helpful and friendly.\n\n"
            "You have specialized team capabilities available through this application:\n"
            "- **Coin Search**: Find coins currently for sale from reputable dealers "
            "(vcoins.com, ma-shops.com, etc.)\n"
            "- **Coin Shows**: Find upcoming coin shows, conventions, and numismatic "
            "events near the user\n"
            "- **Coin Analysis**: Analyze coin images for identification, grading, "
            "and authenticity\n"
            "- **Portfolio Review**: Analyze the user's collection for strengths, "
            "gaps, and recommendations\n\n"
            "If the user's question relates to any of these, let them know they can "
            "ask directly. For example: 'Would you like me to search for those coins?' "
            "or 'I can look up shows near you.'\n\n"
            "For general numismatic questions (history, grading standards, terminology, "
            "collecting tips), answer from your knowledge. Do not use emojis."
        )
        messages = [SystemMessage(content=general_system)] + state["messages"]
        response = await general_model.ainvoke(messages)
        return {"messages": [response]}

    router = create_router(llm_config)

    graph = StateGraph(MessagesState)

    graph.add_node("router", router)
    graph.add_node("coin_search", coin_search_node)
    graph.add_node("coin_shows", coin_shows_node)
    graph.add_node("analysis", analysis_node or passthrough)
    graph.add_node("portfolio", portfolio_node)
    graph.add_node("general", general_handler)

    graph.set_entry_point("router")

    graph.add_edge("coin_search", END)
    graph.add_edge("coin_shows", END)
    graph.add_edge("analysis", END)
    graph.add_edge("portfolio", END)
    graph.add_edge("general", END)

    return graph.compile()
