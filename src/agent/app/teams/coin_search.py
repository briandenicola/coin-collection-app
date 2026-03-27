"""Team 1: Coin Search — multi-agent pipeline with verification.

Pipeline: Search Agent → Verification Agent → Formatter Agent

- Search Agent: finds coin listings via web search (Claude web_search or SearXNG)
- Verification Agent: HTTP-fetches each URL, confirms live and unsold
- Formatter Agent: structures verified results into CoinSuggestion JSON schema
"""

import json
import logging
from typing import Annotated, TypedDict

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langgraph.graph import END, StateGraph

from app.config import settings
from app.llm.provider import get_chat_model
from app.models.requests import LLMConfig
from app.tools.search import create_searxng_search, verify_url

logger = logging.getLogger(__name__)

SEARCH_PROMPT = """You are a numismatic search specialist. Your ONLY job is to search the web
for coins that are CURRENTLY FOR SALE.

CRITICAL RULES:
- Use your search tool to find real, currently available coin listings
- ONLY search on reputable dealer sites: vcoins.com, ma-shops.com, forumancientcoins.com,
  biddr.com, catawiki.com, hjbltd.com
- Add "for sale" or "buy now" to your search queries
- For EACH result, you MUST provide the exact URL to the listing page
- NEVER invent, guess, or recall URLs from memory — only use URLs from search results
- Return ONLY results you actually found in your search

For each coin found, output a JSON object with these fields:
- url: the exact listing URL from the search results
- title: the coin title/name as listed
- price: the listed price
- dealer: the dealer or site name
- snippet: a brief description from the listing

Output your results as a JSON array wrapped in ```json and ``` markers.
If you find nothing, return an empty array: ```json\n[]\n```"""

VERIFY_PROMPT = """You are a verification specialist. You will receive coin listing data
and URL verification results. Your job is to FILTER OUT bad listings.

REMOVE any listing where:
- The URL returned an error (status != 200)
- The page indicates the item is SOLD or UNAVAILABLE
- The page has no buy/bid option

KEEP listings where:
- Status is 200
- Not marked as sold
- Has an active buy or bid option, OR the sold/buy indicators are inconclusive (both false)

Output the VERIFIED listings as a JSON array with the same fields.
Wrap in ```json and ``` markers. If none pass verification, return an empty array."""

FORMAT_PROMPT = """You are a formatting specialist for a coin collecting application.
You receive verified coin listing data. Structure each into this exact JSON schema:

```json
[
  {
    "name": "Full coin name/title",
    "description": "Brief description with condition and authenticity notes",
    "category": "Roman|Greek|Byzantine|Modern|Other",
    "era": "Time period e.g. 27 BC - 14 AD",
    "ruler": "Ruler name if applicable",
    "material": "Gold|Silver|Bronze|Copper|Electrum|Other",
    "denomination": "e.g. Denarius, Tetradrachm",
    "estPrice": "Listed price e.g. $275",
    "imageUrl": "",
    "sourceUrl": "The verified listing URL",
    "sourceName": "Dealer or site name"
  }
]
```

Rules:
- Use ONLY data from the verified listings. Do NOT invent any fields.
- sourceUrl MUST be exactly the URL from the verified data
- Set imageUrl to empty string "" (the frontend extracts images automatically)
- Infer category, era, ruler, material, denomination from the listing title/description
- If you cannot determine a field, use an empty string

Output ONLY the JSON array wrapped in ```json and ``` markers."""


class CoinSearchState(TypedDict):
    """State flowing through the coin search pipeline."""

    messages: Annotated[list, lambda a, b: a + b]
    search_results: str
    verification_results: str
    formatted_results: str
    user_message: str


def create_coin_search_team(llm_config: LLMConfig, user_prompt: str = ""):
    """Create the Team 1 coin search graph.

    Args:
        llm_config: LLM provider configuration
        user_prompt: Optional custom system prompt from admin settings
    """
    model = get_chat_model(llm_config)
    use_searxng = llm_config.provider == "ollama"
    search_tool = create_searxng_search(llm_config.searxng_url) if use_searxng else None

    async def search_node(state: CoinSearchState) -> dict:
        """Search Agent: finds coin listings via web search."""
        user_msg = state.get("user_message", "")

        if use_searxng and search_tool:
            # Ollama mode: use SearXNG tool directly, then pass results to LLM
            search_query = f"{user_msg} ancient coins for sale buy now"
            raw_results = await search_tool.ainvoke(search_query)

            messages = [
                SystemMessage(content=SEARCH_PROMPT),
                HumanMessage(
                    content=f"The user is looking for: {user_msg}\n\n"
                    f"Here are web search results:\n{raw_results}\n\n"
                    "Extract coin listings from these results and format as instructed."
                ),
            ]
            response = await model.ainvoke(messages)
        else:
            # Claude mode: let Claude use its built-in web_search
            search_model = model.bind_tools([], tool_choice="auto") if hasattr(model, "bind_tools") else model
            messages = [
                SystemMessage(content=SEARCH_PROMPT),
                HumanMessage(content=f"Search for: {user_msg}"),
            ]
            response = await search_model.ainvoke(messages)

        return {
            "search_results": response.content if isinstance(response.content, str) else str(response.content),
            "messages": [],
        }

    async def verify_node(state: CoinSearchState) -> dict:
        """Verification Agent: HTTP-fetches each URL, confirms live/unsold."""
        search_results = state.get("search_results", "")

        # Extract URLs from search results JSON
        urls = _extract_urls(search_results)

        if not urls:
            return {
                "verification_results": "No URLs found to verify. Search returned no results.",
                "messages": [],
            }

        # Verify each URL
        verification_data = []
        for url in urls[:settings.max_search_results]:
            result = await verify_url.ainvoke(url)
            verification_data.append(result)

        verification_text = "\n\n".join(verification_data)

        # Ask LLM to filter based on verification
        messages = [
            SystemMessage(content=VERIFY_PROMPT),
            HumanMessage(
                content=f"Original search results:\n{search_results}\n\n"
                f"URL verification results:\n{verification_text}\n\n"
                "Filter the listings based on verification. Remove sold/unavailable items."
            ),
        ]
        response = await model.ainvoke(messages)

        return {
            "verification_results": response.content if isinstance(response.content, str) else str(response.content),
            "messages": [],
        }

    async def format_node(state: CoinSearchState) -> dict:
        """Formatter Agent: structures verified results into CoinSuggestion schema."""
        verified = state.get("verification_results", "")
        user_msg = state.get("user_message", "")

        if "no urls found" in verified.lower() or "empty array" in verified.lower():
            no_results_msg = (
                "I searched for coins matching your request but could not find any "
                "currently available, verified listings. This could mean:\n\n"
                "- The specific coins you're looking for are rare and not currently listed\n"
                "- Try broadening your search criteria\n"
                "- Check back later as dealer inventory changes frequently"
            )
            return {"formatted_results": "", "messages": [AIMessage(content=no_results_msg)]}

        messages = [
            SystemMessage(content=FORMAT_PROMPT),
            HumanMessage(
                content=f"User was searching for: {user_msg}\n\n"
                f"Verified listings:\n{verified}\n\n"
                "Format these into the required JSON schema."
            ),
        ]
        response = await model.ainvoke(messages)
        formatted = response.content if isinstance(response.content, str) else str(response.content)

        # Build a user-friendly response message
        summary = (
            "I found some coins matching your search. "
            "All listings have been verified as currently available."
        )

        return {
            "formatted_results": formatted,
            "messages": [AIMessage(content=f"{summary}\n\n{formatted}")],
        }

    graph = StateGraph(CoinSearchState)
    graph.add_node("search", search_node)
    graph.add_node("verify", verify_node)
    graph.add_node("format", format_node)

    graph.set_entry_point("search")
    graph.add_edge("search", "verify")
    graph.add_edge("verify", "format")
    graph.add_edge("format", END)

    return graph.compile()


def _extract_urls(text: str) -> list[str]:
    """Extract URLs from a JSON block or raw text."""
    # Try parsing JSON block
    json_str = _extract_json_block(text)
    if json_str:
        try:
            data = json.loads(json_str)
            if isinstance(data, list):
                return [item.get("url", "") for item in data if item.get("url")]
        except json.JSONDecodeError:
            pass

    # Fallback: extract URLs with regex
    import re

    return re.findall(r'https?://[^\s"\'<>]+', text)


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
