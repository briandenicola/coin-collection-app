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
    "coinshows-usa.com",
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


# Standard browser user-agent — many dealer sites block bot-like strings
_USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
    "AppleWebKit/537.36 (KHTML, like Gecko) "
    "Chrome/131.0.0.0 Safari/537.36"
)


@tool
async def verify_url(url: str) -> str:
    """Fetch a URL and return its status and key content indicators.

    Used by verification agents to confirm a listing is live and available.
    """
    from urllib.parse import urlparse

    parsed = urlparse(url)
    domain = parsed.netloc.lower().lstrip("www.")
    is_trusted = any(d in domain for d in TRUSTED_DOMAINS)

    # Detect search/category page URLs (not individual listings)
    path_lower = parsed.path.lower() + "?" + (parsed.query or "").lower()
    search_indicators = [
        "search.aspx", "/search?", "/search/", "/browse", "/category/",
        "/results", "viewmode=", "searchterm=", "/en/search",
    ]
    is_search_page = any(ind in path_lower for ind in search_indicators)

    try:
        async with httpx.AsyncClient(timeout=settings.verification_timeout, follow_redirects=True) as client:
            resp = await client.get(url, headers={"User-Agent": _USER_AGENT})

        status = resp.status_code
        text = resp.text[:5000].lower()

        sold_indicators = ["sold", "auction ended", "realized price", "no longer available", "out of stock"]
        is_sold = any(indicator in text for indicator in sold_indicators)

        buy_indicators = ["add to cart", "buy now", "add to basket", "purchase", "bid now", "place bid"]
        has_buy = any(indicator in text for indicator in buy_indicators)

        # Also detect search pages from page content
        if not is_search_page:
            search_content_hints = ["showing results for", "items found", "search results", "refine your search"]
            is_search_page = any(hint in text for hint in search_content_hints)

        return (
            f"Status: {status}\n"
            f"Trusted Dealer Site: {is_trusted}\n"
            f"Search/Category Page (NOT individual listing): {is_search_page}\n"
            f"Sold/Unavailable: {is_sold}\n"
            f"Has Buy/Bid Option: {has_buy}\n"
            f"URL: {url}"
        )
    except Exception as e:
        return (
            f"Error fetching URL: {e}\n"
            f"Trusted Dealer Site: {is_trusted}\n"
            f"Search/Category Page: {is_search_page}\n"
            f"URL: {url}"
        )
