package services

import (
	"testing"
)

func TestShouldHandleCollection(t *testing.T) {
	svc := &CollectionToolsService{}

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		// === Positive cases: should route to collection chat ===

		// Regression: existing explicit collection tokens
		{name: "my collection", message: "Show me my collection", expected: true},
		{name: "my coins", message: "What are my coins worth", expected: true},
		{name: "i own", message: "Which coins do I own from Rome", expected: true},
		{name: "holdings", message: "What are my holdings", expected: true},
		{name: "wishlist", message: "Show my wishlist", expected: true},
		{name: "sold coins", message: "Which sold coins did I have", expected: true},
		{name: "coin id", message: "Show me coin #123", expected: true},
		{name: "how many", message: "How many Roman coins do I have", expected: true},
		{name: "total value", message: "What is the total value of my collection", expected: true},
		{name: "update note", message: "Update the note for coin 5", expected: true},
		{name: "set grade", message: "Set the grade to VF for this coin", expected: true},

		// NEW: ownership-question patterns (DEFECT 1 fix)
		{name: "do i have any X", message: "Do I have any moose coins", expected: true},
		{name: "do i have X and value", message: "Do I have any moose coins and how much are they worth if i have them", expected: true},
		{name: "do i own", message: "Do I own any silver denarii", expected: true},
		{name: "have i got", message: "Have I got any Byzantine coins", expected: true},
		{name: "have i gotten", message: "Have I gotten any new additions this month", expected: true},
		{name: "did i buy", message: "Did I buy any coins from that dealer", expected: true},
		{name: "which of my", message: "Which of my coins are gold", expected: true},
		{name: "show me my X", message: "Show me my Greek coins", expected: true},
		{name: "find in my", message: "Find Nero in my collection", expected: true},
		{name: "any of my", message: "Are any of my coins rare", expected: true},
		{name: "are any of my", message: "Are any of my denarii high grade", expected: true},
		{name: "is there a X", message: "Is there a Constantine coin in my collection", expected: true},
		{name: "are there any X", message: "Are there any moose coins in my collection", expected: true},

		// === Negative cases: should fall through to Python supervisor ===

		// coin_search team (finding coins to BUY)
		{name: "to buy", message: "Find me a Roman denarius to buy", expected: false},
		{name: "for sale", message: "What Roman coins are for sale", expected: false},
		{name: "dealer listings", message: "Show me dealer listings for gold coins", expected: false},
		{name: "vcoins", message: "Search vcoins for Byzantine coins", expected: false},
		{name: "ma-shops", message: "Find Nero coins on ma-shops", expected: false},

		// coin_shows team
		{name: "coin show", message: "Any coin shows near me", expected: false},
		{name: "upcoming show", message: "What upcoming shows are there", expected: false},

		// auction_search team
		{name: "auction", message: "Search auction results for denarii", expected: false},

		// price_trends team (market history, not collection valuation)
		// These do NOT contain ownership framing, so they fall through correctly

		// portfolio team (whole-collection aggregate analysis)
		// "How much are they worth" alone (no ownership context) could go to portfolio,
		// but "do I have X and how much are they worth" contains ownership framing → collection chat

		// Ambiguous/edge cases that should fall through (let Python supervisor decide)
		{name: "generic value question", message: "How much is a denarius worth", expected: false},
		{name: "generic price question", message: "What are Roman coin prices doing this year", expected: false},

		// Empty or whitespace
		{name: "empty", message: "", expected: false},
		{name: "whitespace", message: "   ", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.ShouldHandleCollection(tt.message)
			if result != tt.expected {
				t.Errorf("ShouldHandleCollection(%q) = %v, expected %v", tt.message, result, tt.expected)
			}
		})
	}
}
