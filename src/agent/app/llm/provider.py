"""LLM provider factory — selects Anthropic or Ollama based on request config."""

from langchain_core.language_models import BaseChatModel

from app.models.requests import LLMConfig


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
