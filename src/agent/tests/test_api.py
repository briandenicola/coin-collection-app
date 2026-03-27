"""Basic tests for the agent service."""

from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "ok"
    assert data["service"] == "agent"


def test_search_coins_stub():
    resp = client.post(
        "/api/search/coins",
        json={
            "llm": {"provider": "anthropic", "api_key": "test", "model": "test"},
            "user": {"user_id": 1},
            "message": "Find Roman denarii",
        },
    )
    assert resp.status_code == 200
    data = resp.json()
    assert "message" in data


def test_search_shows_stub():
    resp = client.post(
        "/api/search/shows",
        json={
            "llm": {"provider": "anthropic", "api_key": "test", "model": "test"},
            "user": {"user_id": 1},
            "message": "Upcoming coin shows",
        },
    )
    assert resp.status_code == 200


def test_analyze_stub():
    resp = client.post(
        "/api/analyze",
        json={
            "llm": {"provider": "ollama", "ollama_url": "http://localhost:11434", "model": "llava"},
            "coin": {"id": 1, "name": "Test Coin"},
        },
    )
    assert resp.status_code == 200


def test_portfolio_review_stub():
    resp = client.post(
        "/api/portfolio/review",
        json={
            "llm": {"provider": "anthropic", "api_key": "test", "model": "test"},
            "user": {"user_id": 1},
            "portfolio": {"total_coins": 10, "total_value": 5000},
        },
    )
    assert resp.status_code == 200
