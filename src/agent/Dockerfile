# Build Python agent service
FROM python:3.12-slim@sha256:d764629ce0ddd8c71fd371e9901efb324a95789d2315a47db7e4d27e78f1b0e9 AS builder
WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN pip install --no-cache-dir uv==0.11.22 \
    && uv sync --locked --no-dev --no-install-project

# Final image
FROM python:3.12-slim@sha256:d764629ce0ddd8c71fd371e9901efb324a95789d2315a47db7e4d27e78f1b0e9
WORKDIR /app
COPY --from=builder /app/.venv /app/.venv
COPY app/ ./app/

ENV PYTHONUNBUFFERED=1
ENV PATH="/app/.venv/bin:$PATH"
EXPOSE 8081

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8081"]
