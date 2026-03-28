"""LLM provider factory — selects Anthropic or Ollama based on request config."""

from langchain_core.language_models import BaseChatModel
from langchain_core.runnables import Runnable
from langgraph.graph.state import CompiledStateGraph

from app.models.requests import LLMConfig

# Anthropic server-side tool — executed by Anthropic's servers, not us.
WEB_SEARCH_TOOL = {"type": "web_search_20250305", "name": "web_search"}


def get_chat_model(config: LLMConfig) -> BaseChatModel:
    """Create a LangChain chat model from the per-request LLM config."""
    if config.provider == "anthropic":
        from langchain_anthropic import ChatAnthropic

        return ChatAnthropic(
            model=config.model or "claude-sonnet-4-20250514",
            api_key=config.api_key,
            max_tokens=4096,
        )
    elif config.provider == "ollama":
        from langchain_ollama import ChatOllama

        return ChatOllama(
            model=config.model or "llama3.1",
            base_url=config.ollama_url or "http://localhost:11434",
        )
    else:
        raise ValueError(f"Unknown LLM provider: {config.provider}")


def get_search_model(config: LLMConfig) -> Runnable:
    """Create a chat model with web search enabled (Anthropic only).

    For Anthropic: binds the built-in web_search tool (server-side, handled
    by Anthropic's servers — no local tool execution needed).
    For Ollama: returns a plain model. Use create_search_agent() instead
    for full tool-calling search via SearXNG.
    """
    model = get_chat_model(config)
    if config.provider == "anthropic":
        return model.bind_tools([WEB_SEARCH_TOOL])
    return model


def create_search_agent(config: LLMConfig) -> CompiledStateGraph:
    """Create a ReAct search agent for Ollama with SearXNG tool.

    Returns a compiled LangGraph agent that accepts {"messages": [...]}
    and returns {"messages": [...]}.  The model decides when to call
    the SearXNG search tool, mirroring how Anthropic's built-in
    web_search works server-side.

    Only meaningful for Ollama — Anthropic uses get_search_model() instead.
    """
    from langgraph.prebuilt import create_react_agent

    from app.tools.search import create_searxng_search

    model = get_chat_model(config)
    search_tool = create_searxng_search(config.searxng_url)
    return create_react_agent(model, tools=[search_tool])
