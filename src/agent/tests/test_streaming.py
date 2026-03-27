"""Tests for SSE streaming utilities."""

from app.streaming import extract_suggestions, remove_json_block


def test_extract_suggestions_valid_json():
    text = '''Here are some coins:
```json
[{"name": "Augustus Denarius", "sourceUrl": "https://example.com/1"}]
```
Hope that helps!'''
    result = extract_suggestions(text)
    assert len(result) == 1
    assert result[0]["name"] == "Augustus Denarius"


def test_extract_suggestions_empty_array():
    text = '```json\n[]\n```'
    result = extract_suggestions(text)
    assert result == []


def test_extract_suggestions_no_json():
    text = "I could not find any coins matching your request."
    result = extract_suggestions(text)
    assert result == []


def test_extract_suggestions_invalid_json():
    text = "```json\n{invalid}\n```"
    result = extract_suggestions(text)
    assert result == []


def test_remove_json_block():
    text = '''Found coins:
```json
[{"name": "Test"}]
```
All verified!'''
    result = remove_json_block(text)
    assert "```json" not in result
    assert "Found coins:" in result
    assert "All verified!" in result


def test_remove_json_block_no_block():
    text = "No coins found."
    result = remove_json_block(text)
    assert result == text
