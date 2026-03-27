"""LLM provider factory — selects Anthropic or Ollama based on request config."""

from langchain_core.language_models import BaseChatModel
from langchain_core.runnables import Runnable

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
    """Create a chat model with web search enabled.

    For Anthropic: binds the built-in web_search tool (server-side, handled
    by Anthropic's servers — no local tool execution needed).
    For Ollama: returns a plain model (use SearXNG tool separately).
    """
    model = get_chat_model(config)
    if config.provider == "anthropic":
        return model.bind_tools([WEB_SEARCH_TOOL])
    return model
