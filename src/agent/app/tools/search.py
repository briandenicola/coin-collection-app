"""Web search tool abstraction.

- Anthropic/Claude: uses Claude's built-in web_search (handled by LangChain)
- Ollama: uses SearXNG via HTTP
"""

import httpx
from langchain_core.tools import tool

from app.config import settings

# Trusted coin dealer domains for search filtering
TRUSTED_DOMAINS = [
    "vcoins.com",
    "forumancientcoins.com",
    "hjbltd.com",
    "biddr.com",
    "catawiki.com",
    "ma-shops.com",
    "coinshows.com",
    "money.org",
    "pngdealers.org",
    "nyinc.info",
]


def create_searxng_search(searxng_url: str = ""):
    """Create a SearXNG search tool with a specific URL."""
    url = searxng_url or settings.searxng_url

    @tool
    async def searxng_search(query: str) -> str:
        """Search the web using SearXNG. Used when LLM provider is Ollama."""
        try:
            async with httpx.AsyncClient(timeout=settings.verification_timeout) as client:
                resp = await client.get(
                    f"{url}/search",
                    params={
                        "q": query,
                        "format": "json",
                        "engines": "google,bing,duckduckgo",
                        "categories": "general",
                    },
                )
                resp.raise_for_status()
                data = resp.json()
        except httpx.HTTPError as e:
            return f"Search error: {e}. SearXNG may be unavailable."
        except Exception as e:
            return f"Search failed: {e}"

        results = data.get("results", [])[:settings.max_search_results]
        if not results:
            return "No results found."

        formatted = []
        for r in results:
            formatted.append(f"Title: {r.get('title', '')}\nURL: {r.get('url', '')}\nSnippet: {r.get('content', '')}\n")
        return "\n---\n".join(formatted)

    return searxng_search


# Default instance for backward compatibility
searxng_search = create_searxng_search()


@tool
async def verify_url(url: str) -> str:
    """Fetch a URL and return its status and key content indicators.

    Used by verification agents to confirm a listing is live and available.
    """
    try:
        async with httpx.AsyncClient(timeout=settings.verification_timeout, follow_redirects=True) as client:
            resp = await client.get(url, headers={"User-Agent": "Mozilla/5.0 AncientCoinsBot/1.0"})

        status = resp.status_code
        text = resp.text[:5000].lower()

        sold_indicators = ["sold", "auction ended", "realized price", "no longer available", "out of stock"]
        is_sold = any(indicator in text for indicator in sold_indicators)

        buy_indicators = ["add to cart", "buy now", "add to basket", "purchase", "bid now", "place bid"]
        has_buy = any(indicator in text for indicator in buy_indicators)

        return (
            f"Status: {status}\n"
            f"Sold/Unavailable: {is_sold}\n"
            f"Has Buy/Bid Option: {has_buy}\n"
            f"URL: {url}"
        )
    except Exception as e:
        return f"Error fetching URL: {e}\nURL: {url}"
