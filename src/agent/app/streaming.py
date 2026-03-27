"""SSE streaming utilities for LangGraph agent output.

Converts LangGraph graph events into Server-Sent Events that the Go API
proxies directly to the Vue frontend.
"""

import json
import logging
import re
from collections.abc import AsyncGenerator

from langchain_core.messages import AIMessage, AIMessageChunk

logger = logging.getLogger(__name__)


async def stream_graph_events(graph, input_data: dict, config: dict | None = None) -> AsyncGenerator[str, None]:
    """Stream LangGraph events as SSE-formatted strings.

    Yields SSE data lines compatible with the Go API's existing SSE proxy format:
    - data: {"type": "text", "text": "..."} for incremental text
    - data: {"type": "done", "message": "...", "suggestions": [...]} for completion
    - data: {"type": "error", "message": "..."} on failure
    """
    full_text = ""
    last_ai_content = ""
    last_node_message = ""

    try:
        async for event in graph.astream_events(input_data, config=config or {}, version="v2"):
            kind = event.get("event", "")

            if kind == "on_chat_model_stream":
                chunk = event.get("data", {}).get("chunk")
                if isinstance(chunk, AIMessageChunk) and chunk.content:
                    text = chunk.content if isinstance(chunk.content, str) else ""
                    if text:
                        full_text += text
                        yield format_sse({"type": "text", "text": text})

            elif kind == "on_chat_model_end":
                output = event.get("data", {}).get("output")
                if isinstance(output, AIMessage) and output.content:
                    content = output.content if isinstance(output.content, str) else ""
                    if content:
                        last_ai_content = content

            elif kind == "on_chain_end":
                # Capture AIMessages from node outputs (for paths that skip LLM calls)
                output = event.get("data", {}).get("output", {})
                if isinstance(output, dict):
                    for msg in output.get("messages", []):
                        if isinstance(msg, AIMessage) and msg.content:
                            content = msg.content if isinstance(msg.content, str) else ""
                            if content:
                                last_node_message = content

    except Exception:
        logger.exception("Error during graph streaming")
        yield format_sse({"type": "error", "message": "An error occurred while processing your request."})
        return

    logger.debug(
        "Stream complete — full_text=%d chars, last_ai_content=%d chars, "
        "last_node_message=%d chars",
        len(full_text), len(last_ai_content), len(last_node_message),
    )

    # Build the authoritative final response.
    #
    # Team workflows (coin_search, coin_shows, portfolio) run as sub-graphs
    # via ainvoke(). Their internal LLM events (on_chat_model_stream/end) may
    # NOT propagate to the supervisor's astream_events. The team's final
    # AIMessage IS visible through on_chain_end events on the supervisor node,
    # captured as last_node_message.
    #
    # For direct handler nodes (general_handler), last_ai_content has the
    # correct LLM output AND last_node_message also captures it.
    #
    # Priority: last_node_message > last_ai_content > full_text
    final_text = last_node_message or last_ai_content or full_text
    suggestions = extract_suggestions(final_text)
    clean_message = remove_json_block(final_text) if suggestions else final_text

    # If removing the JSON block left nothing meaningful, provide a fallback
    if not clean_message.strip() and suggestions:
        clean_message = "Here are the results I found."

    done_event: dict = {"type": "done", "message": clean_message.strip()}
    if suggestions:
        done_event["suggestions"] = suggestions

    yield format_sse(done_event)


def format_sse(data: dict) -> str:
    """Format a dict as an SSE data line."""
    return f"data: {json.dumps(data)}\n\n"


def extract_suggestions(text: str) -> list[dict]:
    """Extract coin suggestions from a ```json block in the text."""
    json_str = _extract_json_block(text)
    if not json_str:
        return []
    try:
        data = json.loads(json_str)
        if isinstance(data, list):
            return data
    except json.JSONDecodeError:
        pass
    return []


def remove_json_block(text: str) -> str:
    """Remove the first ```json ... ``` block from text."""
    return re.sub(r"```json\s*\n.*?\n```", "", text, count=1, flags=re.DOTALL)


def _extract_json_block(text: str) -> str | None:
    """Extract the content of the first ```json ... ``` block."""
    start = text.find("```json")
    if start == -1:
        return None
    start += len("```json")
    end = text.find("```", start)
    if end == -1:
        return None
    return text[start:end].strip()
