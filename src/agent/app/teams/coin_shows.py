"""Team 2: Coin Shows — multi-agent pipeline with date verification.

Pipeline: Search Agent → Date Verification Agent → Formatter Agent

- Search Agent: finds upcoming coin shows via web search
- Date Verification Agent: confirms show dates are in the future and not cancelled
- Formatter Agent: structures verified shows into CoinShow JSON schema
"""

import json
import logging
from typing import Annotated, TypedDict

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langgraph.graph import END, StateGraph

from app.llm.provider import get_chat_model
from app.models.requests import LLMConfig, UserContext
from app.tools.search import create_searxng_search

logger = logging.getLogger(__name__)

SEARCH_PROMPT = """You are a coin show search specialist. Your ONLY job is to search for
upcoming coin shows, numismatic conventions, and coin collecting events.

CRITICAL RULES:
- Use your search tool to find REAL, UPCOMING coin shows and events
- Search on reputable sources: coinshows.com, ANA (money.org), PNG (pngdealers.org),
  NYINC, Whitman Expo, local coin club sites, Eventbrite (coin shows)
- Look for shows within the next 12 months
- For EACH result, provide the exact URL from the search results
- NEVER invent, guess, or recall event details from memory
- Return ONLY events you actually found in your search

{location_context}

For each show found, output a JSON object with these fields:
- name: the show/event name
- dates: the show dates as listed (e.g. "March 15-17, 2025")
- location: city/state/country
- venue: venue name if available
- url: the source URL from search results
- description: brief description
- entryFee: entry fee if listed
- dealers: notable dealers or "bourse" info if listed

Output your results as a JSON array wrapped in ```json and ``` markers.
If you find nothing, return an empty array: ```json\n[]\n```"""

DATE_VERIFY_PROMPT = """You are a date verification specialist for coin show events.
You will receive coin show data. Your job is to FILTER OUT invalid shows.

Today's date context will be provided. REMOVE any show where:
- The dates are in the PAST (already happened)
- The show is marked as CANCELLED or POSTPONED INDEFINITELY
- The dates are ambiguous and cannot be confirmed as future
- The URL returned an error or does not match a real event page

KEEP shows where:
- Dates are clearly in the future
- Show appears to be actively scheduled (not cancelled)
- Information comes from a real event page or listing

Output the VERIFIED shows as a JSON array with the same fields.
Wrap in ```json and ``` markers. If none pass verification, return an empty array."""

FORMAT_PROMPT = """You are a formatting specialist for a coin collecting application.
You receive verified coin show data. Structure each into this exact JSON schema:

```json
[
  {
    "name": "Full show name",
    "dates": "Human-readable date range e.g. March 15-17, 2025",
    "location": "City, State/Province, Country",
    "venue": "Venue name",
    "url": "Source URL",
    "description": "Brief description of the show",
    "entryFee": "Entry fee or 'Free' or 'Unknown'",
    "notableDealers": ["Dealer 1", "Dealer 2"]
  }
]
```

Rules:
- Use ONLY data from the verified shows. Do NOT invent any fields.
- url MUST be exactly the URL from the verified data
- Normalize dates into a human-readable format
- Normalize location into "City, State, Country" format
- If a field is unknown, use empty string or empty array
- Sort by date (soonest first)

Output ONLY the JSON array wrapped in ```json and ``` markers."""


class CoinShowSearchState(TypedDict):
    """State flowing through the coin show search pipeline."""

    messages: Annotated[list, lambda a, b: a + b]
    search_results: str
    verification_results: str
    formatted_results: str
    user_message: str
    location_context: str


def create_coin_show_team(
    llm_config: LLMConfig,
    user_context: UserContext | None = None,
    agent_prompt: str = "",
):
    """Create the Team 2 coin show search graph.

    Args:
        llm_config: LLM provider configuration
        user_context: Optional user context with zip code for location
        agent_prompt: Optional custom system prompt from admin settings
    """
    model = get_chat_model(llm_config)
    use_searxng = llm_config.provider == "ollama"
    search_tool = create_searxng_search(llm_config.searxng_url) if use_searxng else None

    # Build location context from user's zip code
    location_hint = ""
    if user_context and user_context.zip_code:
        location_hint = (
            f"The user is located near ZIP code {user_context.zip_code}. "
            "Prioritize shows near this location, but also include major "
            "national shows."
        )

    async def search_node(state: CoinShowSearchState) -> dict:
        """Search Agent: finds upcoming coin shows via web search."""
        user_msg = state.get("user_message", "")
        loc_ctx = state.get("location_context", location_hint)

        prompt = SEARCH_PROMPT.format(
            location_context=loc_ctx if loc_ctx else "No specific location preference."
        )

        if use_searxng and search_tool:
            search_query = f"{user_msg} upcoming coin show numismatic convention"
            if user_context and user_context.zip_code:
                search_query += f" near {user_context.zip_code}"
            raw_results = await search_tool.ainvoke(search_query)

            messages = [
                SystemMessage(content=prompt),
                HumanMessage(
                    content=f"The user asked: {user_msg}\n\n"
                    f"Here are web search results:\n{raw_results}\n\n"
                    "Extract coin shows from these results and format as instructed."
                ),
            ]
            response = await model.ainvoke(messages)
        else:
            # Claude mode: use built-in web_search
            messages = [
                SystemMessage(content=prompt),
                HumanMessage(content=f"Search for: {user_msg}"),
            ]
            response = await model.ainvoke(messages)

        return {
            "search_results": response.content if isinstance(response.content, str) else str(response.content),
            "messages": [],
        }

    async def verify_node(state: CoinShowSearchState) -> dict:
        """Date Verification Agent: confirms show dates are future and not cancelled."""
        search_results = state.get("search_results", "")

        if not search_results or "[]" in search_results:
            return {
                "verification_results": "No shows found to verify.",
                "messages": [],
            }

        from datetime import datetime, timezone

        today = datetime.now(tz=timezone.utc).strftime("%B %d, %Y")

        messages = [
            SystemMessage(content=DATE_VERIFY_PROMPT),
            HumanMessage(
                content=f"Today's date is: {today}\n\n"
                f"Coin show search results:\n{search_results}\n\n"
                "Verify that each show's dates are in the future and the show is not cancelled."
            ),
        ]
        response = await model.ainvoke(messages)

        return {
            "verification_results": response.content if isinstance(response.content, str) else str(response.content),
            "messages": [],
        }

    async def format_node(state: CoinShowSearchState) -> dict:
        """Formatter Agent: structures verified shows into CoinShow schema."""
        verified = state.get("verification_results", "")
        user_msg = state.get("user_message", "")

        if "no shows found" in verified.lower() or "empty array" in verified.lower():
            no_results_msg = (
                "I searched for upcoming coin shows but could not find any "
                "verified events matching your criteria. Try:\n\n"
                "- Broadening your search area\n"
                "- Searching for a different time period\n"
                "- Checking coinshows.com or money.org directly"
            )
            return {"formatted_results": "", "messages": [AIMessage(content=no_results_msg)]}

        messages = [
            SystemMessage(content=FORMAT_PROMPT),
            HumanMessage(
                content=f"User was searching for: {user_msg}\n\n"
                f"Verified shows:\n{verified}\n\n"
                "Format these into the required JSON schema."
            ),
        ]
        response = await model.ainvoke(messages)
        formatted = response.content if isinstance(response.content, str) else str(response.content)

        count = _count_shows(formatted)
        summary = (
            f"I found {count} upcoming coin show{'s' if count != 1 else ''} "
            "matching your search. All dates have been verified as upcoming events."
        )

        return {
            "formatted_results": formatted,
            "messages": [AIMessage(content=f"{summary}\n\n{formatted}")],
        }

    graph = StateGraph(CoinShowSearchState)
    graph.add_node("search", search_node)
    graph.add_node("verify", verify_node)
    graph.add_node("format", format_node)

    graph.set_entry_point("search")
    graph.add_edge("search", "verify")
    graph.add_edge("verify", "format")
    graph.add_edge("format", END)

    return graph.compile()


def _count_shows(text: str) -> int:
    """Count the number of shows in a JSON block."""
    json_str = _extract_json_block(text)
    if json_str:
        try:
            data = json.loads(json_str)
            if isinstance(data, list):
                return len(data)
        except json.JSONDecodeError:
            pass
    return 0


def _extract_json_block(text: str) -> str | None:
    """Extract the first ```json ... ``` block from text."""
    start = text.find("```json")
    if start == -1:
        return None
    start += len("```json")
    end = text.find("```", start)
    if end == -1:
        return None
    return text[start:end].strip()
