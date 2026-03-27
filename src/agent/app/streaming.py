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
                    if content and not full_text:
                        full_text = content

    except Exception:
        logger.exception("Error during graph streaming")
        yield format_sse({"type": "error", "message": "An error occurred while processing your request."})
        return

    # Extract suggestions JSON from the final text and strip it from the display message
    suggestions = extract_suggestions(full_text)
    clean_message = remove_json_block(full_text) if suggestions else full_text

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
