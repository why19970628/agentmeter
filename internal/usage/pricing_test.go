package usage

import "testing"

func TestEstimateCostUsesBuiltInModelCatalog(t *testing.T) {
	event := Event{
		ModelName:    "gpt-5",
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
	}

	got := EstimateCost(event)

	if got <= 0 {
		t.Fatalf("cost = %f, want positive estimate", got)
	}
}
