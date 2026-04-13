"""Tests for the availability check endpoint and team pipeline."""

from fastapi.testclient import TestClient

from app.main import app
from app.models.responses import AvailabilityVerdict
from app.teams.availability_check import parse_verdicts

client = TestClient(app)


def test_check_availability_rejects_invalid_body():
    resp = client.post("/api/check-availability", json={})
    assert resp.status_code == 422


def test_check_availability_returns_empty_for_no_items():
    resp = client.post(
        "/api/check-availability",
        json={
            "llm": {"provider": "anthropic", "api_key": "k", "model": "m"},
            "items": [],
        },
    )
    assert resp.status_code == 200
    data = resp.json()
    assert data["results"] == []


def test_parse_verdicts_valid_json():
    raw = '''Here are the results:
```json
[
  {
    "url": "https://example.com/coin/1",
    "coin_name": "Roman Denarius",
    "status": "available",
    "reason": "Buy button found",
    "confidence": "high"
  },
  {
    "url": "https://example.com/coin/2",
    "coin_name": "Greek Tetradrachm",
    "status": "unavailable",
    "reason": "Page shows sold indicator",
    "confidence": "high"
  }
]
```'''
    verdicts = parse_verdicts(raw)
    assert len(verdicts) == 2
    assert verdicts[0].status == "available"
    assert verdicts[0].coin_name == "Roman Denarius"
    assert verdicts[1].status == "unavailable"
    assert verdicts[1].confidence == "high"


def test_parse_verdicts_invalid_json():
    verdicts = parse_verdicts("This is not JSON at all")
    assert verdicts == []


def test_parse_verdicts_no_code_fence():
    raw = '[{"url": "https://x.com", "coin_name": "Test", "status": "unknown", "reason": "ambiguous"}]'
    verdicts = parse_verdicts(raw)
    assert len(verdicts) == 1
    assert verdicts[0].status == "unknown"


def test_availability_verdict_model():
    v = AvailabilityVerdict(
        url="https://example.com",
        status="available",
        reason="Active listing",
    )
    assert v.confidence == "medium"  # default
    assert v.coin_name == ""  # default
