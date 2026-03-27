"""SSE streaming utilities for LangGraph agent output.

Converts LangGraph graph events into Server-Sent Events that the Go API
proxies directly to the Vue frontend.
"""

import json
from collections.abc import AsyncGenerator

from langchain_core.messages import AIMessage, AIMessageChunk


async def stream_graph_events(graph, input_data: dict, config: dict | None = None) -> AsyncGenerator[str, None]:
    """Stream LangGraph events as SSE-formatted strings.

    Yields SSE data lines compatible with the Go API's existing SSE proxy format:
    - data: {"type": "text", "text": "..."} for incremental text
    - data: {"type": "done", "message": "...", "suggestions": [...]} for completion
    """
    full_text = ""

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

    yield format_sse({"type": "done", "message": full_text})


def format_sse(data: dict) -> str:
    """Format a dict as an SSE data line."""
    return f"data: {json.dumps(data)}\n\n"
