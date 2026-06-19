# Python Locking Skill

Use this when changing Python agent dependencies, CI installs, or Docker dependency installation.

1. Keep `src/agent/pyproject.toml` as the dependency manifest and `src/agent/uv.lock` as the committed lock file.
2. Use the reviewed uv pin from workflows/docs. Install with `pip install uv==<pin>`.
3. Refresh intentionally from `src/agent` with `uv lock --upgrade`, then verify with `uv sync --locked --extra dev`.
4. CI lint/tests should run through `uv run` after `uv sync --locked --extra dev`.
5. Docker runtime installs should use `uv sync --locked --no-dev --no-install-project` and copy the resulting virtual environment into the final image.
6. Dependabot should use `package-ecosystem: uv` for `/src/agent` so lock updates are reviewed in PRs.
7. Validate `uv run ruff check app/ tests/`, `uv run pytest tests/ -v`, and an agent Docker build when Docker is available.
