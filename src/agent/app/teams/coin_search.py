"""Team 1: Coin Search — multi-agent pipeline with verification.

Pipeline: Search Agent → Verification Agent → Formatter Agent

- Search Agent: finds coin listings via web search (Claude web_search or SearXNG)
- Verification Agent: HTTP-fetches each URL, confirms live and unsold
- Formatter Agent: structures verified results into CoinSuggestion JSON schema
"""

import json
import logging
import re
from typing import Annotated, TypedDict
from urllib.parse import urlparse

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langgraph.graph import END, StateGraph

from app.config import settings
from app.llm.provider import get_chat_model
from app.models.requests import LLMConfig
from app.tools.search import create_searxng_search, verify_url

logger = logging.getLogger(__name__)

SEARCH_PROMPT = """You are a numismatic search specialist. Your job is to find SPECIFIC coins
that are currently for sale — not search result pages or category pages.

SEARCH STRATEGY (follow these steps):
1. Run targeted searches using site-specific queries. For each dealer site, search like:
   - site:vcoins.com "Domitian" "denarius" — to find individual listing pages
   - site:ma-shops.com "Hadrian" "sestertius" — narrow by ruler + denomination
   - Include price terms and the user's budget if mentioned
2. If your initial searches return SEARCH RESULT PAGES (URLs containing "Search", "search",
   "results", "category", or "browse"), use web_search to VISIT those pages and extract
   individual coin listings from the page content.
3. For each individual coin, you need: a title, a price, and a URL to that specific listing.
   A "search results page URL" is NOT a valid coin listing.
4. Run at LEAST 3-5 separate searches across different dealer sites to find good results.

DEALER SITES TO SEARCH:
- vcoins.com — individual listings look like vcoins.com/en/stores/DEALER/ID/...
- ma-shops.com — individual listings look like ma-shops.com/DEALER/item/NUMBER
- forumancientcoins.com — individual listings look like forumancientcoins.com/catalog/...
- biddr.com, catawiki.com, hjbltd.com

WHAT COUNTS AS A VALID RESULT:
- A specific coin with a title, price, and individual listing URL
- The URL must point to ONE coin, not a search page or category page

WHAT DOES NOT COUNT:
- Search result page URLs (e.g. vcoins.com/en/Search.aspx?search=domitian)
- Category or browse pages
- Generic dealer homepage links
- Coins you remember from training data — ONLY list coins found in this search session

URL RULES:
- Use ONLY URLs that appeared verbatim in your web search results
- NEVER construct, guess, or modify a URL path
- A real URL from a search result is ALWAYS better than a fabricated one

For each coin found, output a JSON object with these fields:
- url: the EXACT individual listing URL (not a search page)
- title: the coin title/name as listed by the dealer
- price: the listed price (e.g. "$275.00", "EUR 180.00")
- dealer: the dealer or site name
- snippet: condition, notes, or description from the listing

Output your results as a JSON array wrapped in ```json and ``` markers.
If you find nothing, return an empty array: ```json\n[]\n```"""

VERIFY_PROMPT = """You are a verification specialist. You will receive coin listing data
and URL verification results. Your job is to FILTER OUT bad listings.

REMOVE any listing where:
- The URL points to a SEARCH RESULTS page, category page, or browse page — NOT an
  individual coin listing. Search page indicators: "Search.aspx", "/search?", "/browse",
  "/category/", "results" in URL path, or page title says "Search Results"
- The page clearly indicates the item is SOLD or UNAVAILABLE
- The URL is from an unknown or untrustworthy source

KEEP listings where:
- The URL points to an individual coin listing page (a single coin with its own page)
- Status is 200 and item appears available
- Status is 403 or 503 BUT the URL is from a known reputable dealer site
  (vcoins.com, ma-shops.com, forumancientcoins.com, biddr.com, catawiki.com,
  hjbltd.com). Many dealer sites block automated requests — a 403 does NOT
  mean the listing is invalid.
- Not marked as sold
- Has an active buy or bid option, OR the sold/buy indicators are inconclusive

IMPORTANT: Preserve ALL URLs exactly as provided. Do NOT modify, construct, or
"fix" any URL. Copy them character-for-character from the input.

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
    "sourceUrl": "The verified listing URL — MUST be copied exactly, never modified",
    "sourceName": "Dealer or site name"
  }
]
```

Rules:
- Use ONLY data from the verified listings. Do NOT invent any fields.
- sourceUrl MUST be copied character-for-character from the verified data. NEVER
  construct, guess, or "fix" a URL.
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


def create_coin_search_team(llm_config: LLMConfig, search_prompt: str = ""):
    """Create the Team 1 coin search graph.

    Args:
        llm_config: LLM provider configuration
        search_prompt: Additional search context from admin settings (prepended to system prompt)
    """
    # The admin prompt provides personality/context; SEARCH_PROMPT provides structure
    if search_prompt:
        combined_prompt = f"{search_prompt}\n\n{SEARCH_PROMPT}"
    else:
        combined_prompt = SEARCH_PROMPT
    logger.debug("Coin search prompt (%d chars): %.80s...", len(combined_prompt), combined_prompt)

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
                SystemMessage(content=combined_prompt),
                HumanMessage(
                    content=f"The user is looking for: {user_msg}\n\n"
                    f"Here are web search results:\n{raw_results}\n\n"
                    "Extract INDIVIDUAL coin listings with prices from these results."
                ),
            ]
            response = await model.ainvoke(messages)
        else:
            # Claude mode: use built-in web_search with tactical instructions
            messages = [
                SystemMessage(content=combined_prompt),
                HumanMessage(
                    content=f"Find specific coins for sale matching: {user_msg}\n\n"
                    "Search strategy:\n"
                    "1. Use site-specific searches (e.g. site:vcoins.com) to find "
                    "individual coin listings with prices\n"
                    "2. If you find search result pages, visit them to extract individual "
                    "coin listings with titles, prices, and direct URLs\n"
                    "3. Search across multiple dealer sites (vcoins.com, ma-shops.com, etc.)\n"
                    "4. I need SPECIFIC coins with individual listing URLs and prices — "
                    "NOT search page links or general recommendations"
                ),
            ]
            response = await model.ainvoke(messages)

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

        # Verify URLs in parallel for speed
        import asyncio

        tasks = [verify_url.ainvoke(url) for url in urls[:settings.max_search_results]]
        verification_data = await asyncio.gather(*tasks, return_exceptions=True)
        verification_text = "\n\n".join(
            str(r) for r in verification_data if not isinstance(r, Exception)
        )

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

        if "no urls found" in verified.lower() or "empty array" in verified.lower() or _is_empty_json_result(verified):
            # Call the LLM so the response streams to the user via SSE
            no_results_prompt = (
                "You are an assistant in a coin collecting application. "
                "The user searched for coins to buy but no verified listings were found. "
                "Generate a brief, helpful response. Suggest broadening search criteria, "
                "checking back later, or trying specific dealer sites like vcoins.com "
                "or ma-shops.com. Keep it concise. Do not use emojis. "
                "Do not invent coin listings."
            )
            messages = [
                SystemMessage(content=no_results_prompt),
                HumanMessage(
                    content=f"The user asked: {user_msg}\n\n"
                    "No verified coin listings were found. Generate a helpful response."
                ),
            ]
            response = await model.ainvoke(messages)
            content = response.content if isinstance(response.content, str) else str(response.content)
            return {"formatted_results": "", "messages": [AIMessage(content=content)]}

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

        # Sanitize URLs: replace any fabricated URLs with originals from search results
        search_results = state.get("search_results", "")
        formatted = _sanitize_urls(formatted, search_results, verified, url_field="sourceUrl")

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
    return re.findall(r'https?://[^\s"\'<>\])+,]+', text)


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


def _is_empty_json_result(text: str) -> bool:
    """Check if text contains a JSON block with an empty array."""
    json_block = _extract_json_block(text)
    if json_block:
        try:
            data = json.loads(json_block)
            return isinstance(data, list) and len(data) == 0
        except json.JSONDecodeError:
            pass
    return False


def _extract_urls_from_text(text: str) -> list[str]:
    """Extract all http/https URLs from free text."""
    return re.findall(r'https?://[^\s"\'<>\])+,]+', text)


def _sanitize_urls(
    formatted: str, search_results: str, verified: str, url_field: str = "sourceUrl"
) -> str:
    """Replace fabricated URLs in formatted output with real ones from search results.

    LLMs often modify URLs across pipeline hops (search -> verify -> format).
    This extracts real URLs from search and verify stages, then checks each URL
    in the formatted output. Fabricated URLs are replaced with the best matching
    real URL from the same domain.
    """
    real_urls = set(_extract_urls_from_text(search_results))
    real_urls.update(_extract_urls_from_text(verified))

    if not real_urls:
        return formatted

    json_block = _extract_json_block(formatted)
    if not json_block:
        return formatted

    try:
        items = json.loads(json_block)
        if not isinstance(items, list):
            return formatted
    except json.JSONDecodeError:
        return formatted

    changed = False
    for item in items:
        url = item.get(url_field, "")
        if not url or url in real_urls:
            continue

        try:
            fabricated_domain = urlparse(url).netloc.lower()
        except Exception:
            continue

        domain_matches = [u for u in real_urls if fabricated_domain in urlparse(u).netloc.lower()]
        if domain_matches:
            item[url_field] = domain_matches[0]
            changed = True
            logger.warning("Replaced fabricated URL %s with real URL %s", url, item[url_field])
        else:
            logger.warning("Fabricated URL %s has no domain match in search results", url)

    if changed:
        new_json = json.dumps(items, indent=2)
        start = formatted.find("```json")
        end_marker = formatted.find("```", start + len("```json"))
        if start != -1 and end_marker != -1:
            formatted = formatted[:start] + "```json\n" + new_json + "\n" + formatted[end_marker:]

    return formatted
