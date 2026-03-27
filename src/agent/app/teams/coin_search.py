"""Team 1: Coin Search — single agent with tools.

Single agent design: Claude searches the web, then uses fetch_dealer_page
to visit dealer search result pages and extract REAL individual listings.
This avoids the fundamental problem of the old 3-agent pipeline where
URLs were fabricated across pipeline hops.
"""

import logging
from typing import Annotated, TypedDict

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langgraph.graph import END, StateGraph

from app.llm.provider import get_chat_model
from app.models.requests import LLMConfig
from app.tools.search import create_searxng_search, fetch_dealer_page

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = """You are a numismatic search assistant helping collectors find coins for sale.

You have access to the `fetch_dealer_page` tool. Here is your workflow:

STEP 1 — SEARCH: Use web_search to find dealer pages selling the coins the user wants.
Search on: vcoins.com, ma-shops.com, forumancientcoins.com, biddr.com, catawiki.com.
Use targeted queries like: "Aurelian antoninianus for sale site:vcoins.com"

STEP 2 — FETCH: When you find dealer search result pages or listing pages, call
`fetch_dealer_page` with that URL. This tool fetches the actual page and extracts
real coin listings with titles, prices, and URLs.

STEP 3 — PRESENT: Format the REAL listings you found into a JSON array.

CRITICAL RULES:
- ALWAYS use fetch_dealer_page to get real listing data — never guess listing URLs
- Only include coins that were found by your tools — never invent listings from memory
- Include the user's budget/price range in your search queries
- Do not use emojis

OUTPUT FORMAT: After gathering listings, output a JSON array wrapped in ```json and ```
markers with these fields for each coin:

```json
[
  {
    "name": "Coin title from the dealer listing",
    "description": "Brief description from the listing",
    "category": "Roman|Greek|Byzantine|Modern|Other",
    "era": "Time period",
    "ruler": "Ruler name",
    "material": "Gold|Silver|Bronze|Copper|Other",
    "denomination": "e.g. Denarius, Antoninianus",
    "estPrice": "Listed price e.g. $150.00",
    "imageUrl": "",
    "sourceUrl": "The REAL URL from fetch_dealer_page — never fabricate",
    "sourceName": "Dealer or site name"
  }
]
```

If you find nothing after searching, say so honestly. Do not invent results."""


class CoinSearchState(TypedDict):
    """State for the coin search agent."""

    messages: Annotated[list, lambda a, b: a + b]
    user_message: str


def create_coin_search_team(llm_config: LLMConfig, search_prompt: str = ""):
    """Create the coin search agent with tools.

    Args:
        llm_config: LLM provider configuration
        search_prompt: Additional context from admin settings (prepended)
    """
    if search_prompt:
        combined_prompt = f"{search_prompt}\n\n{SYSTEM_PROMPT}"
    else:
        combined_prompt = SYSTEM_PROMPT
    logger.debug(
        "Coin search prompt (%d chars): %.80s...",
        len(combined_prompt), combined_prompt,
    )

    use_searxng = llm_config.provider == "ollama"

    if use_searxng:
        # Ollama mode: provide both tools explicitly
        search_tool = create_searxng_search(llm_config.searxng_url)
        tools = [search_tool, fetch_dealer_page]
        model = get_chat_model(llm_config).bind_tools(tools)
    else:
        # Anthropic mode: Claude has built-in web_search + our fetch tool
        model = get_chat_model(llm_config).bind_tools([fetch_dealer_page])

    async def search_agent(state: CoinSearchState) -> dict:
        """Single agent that searches, fetches dealer pages, and formats results."""
        user_msg = state.get("user_message", "")

        messages = [
            SystemMessage(content=combined_prompt),
            HumanMessage(
                content=f"Find coins for sale matching: {user_msg}\n\n"
                "Remember: search first, then use fetch_dealer_page on any "
                "dealer search result pages to get real individual listings."
            ),
        ]

        # Let the agent run with tool calls — loop until it stops calling tools
        max_iterations = 8
        for _ in range(max_iterations):
            response = await model.ainvoke(messages)
            messages.append(response)

            # If no tool calls, the agent is done
            if not response.tool_calls:
                break

            # Execute tool calls
            from langchain_core.messages import ToolMessage

            for tc in response.tool_calls:
                tool_name = tc["name"]
                tool_args = tc["args"]

                if tool_name == "fetch_dealer_page":
                    result = await fetch_dealer_page.ainvoke(tool_args)
                elif use_searxng and tool_name == "searxng_search":
                    result = await search_tool.ainvoke(tool_args)
                else:
                    result = f"Unknown tool: {tool_name}"

                messages.append(ToolMessage(
                    content=str(result),
                    tool_call_id=tc["id"],
                ))

        # Extract final response
        final = response.content if isinstance(response.content, str) else str(response.content)

        return {"messages": [AIMessage(content=final)]}

    graph = StateGraph(CoinSearchState)
    graph.add_node("search_agent", search_agent)
    graph.set_entry_point("search_agent")
    graph.add_edge("search_agent", END)

    return graph.compile()
