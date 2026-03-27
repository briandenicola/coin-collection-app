"""Team 1: Coin Search — two-phase search with page fetching.

Phase 1: Claude searches the web using web_search (enabled via get_search_model).
Phase 2: We fetch dealer pages from the URLs found and extract real listings.
Phase 3: Claude formats the extracted listings into the CoinSuggestion JSON schema.
"""

import logging
import re
from typing import Annotated, TypedDict

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langgraph.graph import END, StateGraph

from app.llm.provider import get_chat_model, get_search_model
from app.models.requests import LLMConfig
from app.tools.search import create_searxng_search, fetch_dealer_page

logger = logging.getLogger(__name__)

SEARCH_PROMPT = """You are a numismatic search specialist. Search the web to find coins
currently for sale that match the user's request.

Search on these dealer sites:
- vcoins.com, ma-shops.com, forumancientcoins.com, biddr.com, catawiki.com, hjbltd.com

Use targeted site-specific queries like:
- "Domitian denarius for sale site:vcoins.com"
- "Greek tetradrachm Athens site:ma-shops.com"

Include the user's budget/price range in your searches if mentioned.
Run at least 3-5 searches across different dealer sites.

For each result you find, report the URL exactly as it appeared in the search results.
Do NOT invent or modify URLs. Do not use emojis."""

FORMAT_PROMPT = """You are a formatting specialist for a coin collecting application.
You receive raw coin listing data extracted from dealer websites.
Structure each listing into this exact JSON schema:

```json
[
  {
    "name": "Coin title from the listing",
    "description": "Brief description from the listing",
    "category": "Roman|Greek|Byzantine|Modern|Other",
    "era": "Time period",
    "ruler": "Ruler name",
    "material": "Gold|Silver|Bronze|Copper|Other",
    "denomination": "e.g. Denarius, Tetradrachm",
    "estPrice": "Listed price e.g. $150.00",
    "imageUrl": "",
    "sourceUrl": "The exact URL from the listing data — never fabricate",
    "sourceName": "Dealer or site name"
  }
]
```

Rules:
- Use ONLY data from the listing extracts. Do NOT invent fields.
- sourceUrl MUST be copied exactly from the data. NEVER fabricate URLs.
- Set imageUrl to "" (the frontend handles images)
- Infer category, era, ruler, material, denomination from the listing text
- If you cannot determine a field, use an empty string
- Do not use emojis

Output ONLY the JSON array wrapped in ```json and ``` markers."""

NO_RESULTS_PROMPT = (
    "You are an assistant in a coin collecting application. "
    "The user searched for coins to buy but no listings were found. "
    "Generate a brief, helpful response. Suggest broadening search criteria, "
    "checking back later, or trying specific dealer sites like vcoins.com "
    "or ma-shops.com. Keep it concise. Do not use emojis. "
    "Do not invent coin listings."
)


class CoinSearchState(TypedDict):
    """State for the coin search pipeline."""

    messages: Annotated[list, lambda a, b: a + b]
    search_results: str
    fetched_listings: str
    user_message: str


def create_coin_search_team(llm_config: LLMConfig, search_prompt: str = ""):
    """Create the coin search pipeline.

    Args:
        llm_config: LLM provider configuration
        search_prompt: Additional context from admin settings (prepended)
    """
    if search_prompt:
        combined_search = f"{search_prompt}\n\n{SEARCH_PROMPT}"
    else:
        combined_search = SEARCH_PROMPT

    use_searxng = llm_config.provider == "ollama"

    async def search_node(state: CoinSearchState) -> dict:
        """Phase 1: Search the web for dealer pages."""
        user_msg = state.get("user_message", "")
        # get_search_model enables Anthropic's built-in web_search tool
        model = get_search_model(llm_config)

        if use_searxng:
            search_tool = create_searxng_search(llm_config.searxng_url)
            query = f"{user_msg} ancient coins for sale buy now"
            raw_results = await search_tool.ainvoke(query)
            messages = [
                SystemMessage(content=combined_search),
                HumanMessage(
                    content=f"The user is looking for: {user_msg}\n\n"
                    f"Here are web search results:\n{raw_results}\n\n"
                    "List the most promising URLs to check for individual listings."
                ),
            ]
        else:
            messages = [
                SystemMessage(content=combined_search),
                HumanMessage(
                    content=f"Find coins for sale matching: {user_msg}\n\n"
                    "Search multiple dealer sites and report all URLs you find."
                ),
            ]

        response = await model.ainvoke(messages)
        content = response.content if isinstance(response.content, str) else str(response.content)
        return {"search_results": content, "messages": []}

    async def fetch_node(state: CoinSearchState) -> dict:
        """Phase 2: Fetch dealer pages and extract real listings."""
        import asyncio

        search_results = state.get("search_results", "")
        urls = _extract_urls(search_results)

        if not urls:
            return {"fetched_listings": "", "messages": []}

        # Fetch up to 5 URLs in parallel
        tasks = [fetch_dealer_page.ainvoke({"url": u}) for u in urls[:5]]
        results = await asyncio.gather(*tasks, return_exceptions=True)

        fetched = []
        for url, result in zip(urls[:5], results):
            if isinstance(result, Exception):
                logger.warning("Failed to fetch %s: %s", url, result)
                continue
            text = str(result)
            if not text.startswith("Error"):
                fetched.append(f"--- Source: {url} ---\n{text}")

        return {"fetched_listings": "\n\n".join(fetched), "messages": []}

    async def format_node(state: CoinSearchState) -> dict:
        """Phase 3: Format extracted listings into CoinSuggestion JSON."""
        fetched = state.get("fetched_listings", "")
        user_msg = state.get("user_message", "")
        search_results = state.get("search_results", "")
        model = get_chat_model(llm_config)

        if not fetched.strip():
            # No listings found — generate a helpful response via LLM (streams)
            messages = [
                SystemMessage(content=NO_RESULTS_PROMPT),
                HumanMessage(
                    content=f"The user asked: {user_msg}\n\n"
                    f"Search results summary:\n{search_results[:1000]}\n\n"
                    "No coin listings could be extracted. Generate a helpful response."
                ),
            ]
            response = await model.ainvoke(messages)
            content = response.content if isinstance(response.content, str) else str(response.content)
            return {"messages": [AIMessage(content=content)]}

        # Format real listings via LLM (this call streams to user)
        messages = [
            SystemMessage(content=FORMAT_PROMPT),
            HumanMessage(
                content=f"User searched for: {user_msg}\n\n"
                f"Extracted listing data:\n{fetched}"
            ),
        ]
        response = await model.ainvoke(messages)
        formatted = response.content if isinstance(response.content, str) else str(response.content)

        summary = (
            "I found some coins matching your search. "
            "Here are the listings I extracted from dealer sites."
        )
        return {"messages": [AIMessage(content=f"{summary}\n\n{formatted}")]}

    graph = StateGraph(CoinSearchState)
    graph.add_node("search", search_node)
    graph.add_node("fetch", fetch_node)
    graph.add_node("format", format_node)

    graph.set_entry_point("search")
    graph.add_edge("search", "fetch")
    graph.add_edge("fetch", "format")
    graph.add_edge("format", END)

    return graph.compile()


def _extract_urls(text: str) -> list[str]:
    """Extract dealer URLs from search results text."""
    urls = re.findall(r'https?://[^\s"\'<>\])+,]+', text)
    # Deduplicate while preserving order
    seen = set()
    unique = []
    for url in urls:
        if url not in seen:
            seen.add(url)
            unique.append(url)
    return unique
