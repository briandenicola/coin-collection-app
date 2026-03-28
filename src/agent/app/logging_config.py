"""Structured logging with in-memory ring buffer.

Mirrors the Go API's logger: writes to stdout AND stores entries in a
ring buffer that can be read via the /admin/logs endpoint. The Go API
merges these into the admin UI.
"""

import logging
import sys
import threading
from collections import deque
from datetime import datetime, timezone


class LogEntry:
    """A single log entry matching the Go LogEntry JSON schema."""

    __slots__ = ("timestamp", "level", "message")

    def __init__(self, timestamp: str, level: str, message: str):
        self.timestamp = timestamp
        self.level = level
        self.message = message

    def to_dict(self) -> dict:
        return {"timestamp": self.timestamp, "level": self.level, "message": self.message}


class RingBufferHandler(logging.Handler):
    """Logging handler that stores entries in a thread-safe ring buffer."""

    def __init__(self, capacity: int = 500):
        super().__init__()
        self._buffer: deque[LogEntry] = deque(maxlen=capacity)
        self._lock = threading.Lock()

    def emit(self, record: logging.LogRecord) -> None:
        entry = LogEntry(
            timestamp=datetime.now(tz=timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            level=record.levelname,
            message=f"[{record.name}] {self.format(record)}",
        )
        with self._lock:
            self._buffer.append(entry)

    def get_logs(self, limit: int = 500, level: str = "") -> list[dict]:
        """Return recent log entries, optionally filtered by level."""
        with self._lock:
            entries = list(self._buffer)
        if level:
            entries = [e for e in entries if e.level == level.upper()]
        if limit > 0:
            entries = entries[-limit:]
        return [e.to_dict() for e in entries]


# Singleton ring buffer handler — shared across the app
ring_handler = RingBufferHandler(capacity=500)


def setup_logging(log_level: str = "INFO") -> None:
    """Configure structured logging for the entire Python agent service.

    - All loggers write to stdout (for docker logs / container output)
    - All loggers also write to the ring buffer (for admin UI)
    - Format matches Go's style: [LEVEL] [component] message
    """
    level = getattr(logging, log_level.upper(), logging.INFO)

    # Structured format matching Go: [LEVEL] [component] message
    formatter = logging.Formatter(
        fmt="%(asctime)s [%(levelname)s] [%(name)s] %(message)s",
        datefmt="%Y-%m-%dT%H:%M:%SZ",
    )

    # Stdout handler
    stdout_handler = logging.StreamHandler(sys.stdout)
    stdout_handler.setFormatter(formatter)

    # Ring buffer handler
    ring_handler.setFormatter(logging.Formatter("%(message)s"))

    # Configure root logger
    root = logging.getLogger()
    root.setLevel(level)
    root.handlers.clear()
    root.addHandler(stdout_handler)
    root.addHandler(ring_handler)

    # Quiet noisy third-party loggers
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("httpcore").setLevel(logging.WARNING)
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)
    logging.getLogger("langchain_core").setLevel(logging.WARNING)
    logging.getLogger("langsmith").setLevel(logging.WARNING)


def set_log_level(log_level: str) -> str:
    """Dynamically change the root logger level. Returns the applied level."""
    level = getattr(logging, log_level.upper(), None)
    if level is None:
        return logging.getLogger().level  # unchanged
    logging.getLogger().setLevel(level)
    return logging.getLevelName(level)
