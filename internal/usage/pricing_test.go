package usage

import (
	"math"
	"testing"
)

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

func TestEstimateCostPricesCachedInputSeparately(t *testing.T) {
	event := Event{
		ToolName:         "openai-api",
		ModelName:        "gpt-5.5",
		InputTokens:      1_000_000,
		OutputTokens:     100_000,
		CacheReadTokens:  900_000,
		CacheWriteTokens: 0,
	}

	got := EstimateCost(event)
	want := 100_000.0/1_000_000*5.00 +
		900_000.0/1_000_000*0.50 +
		100_000.0/1_000_000*30.00

	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("cost = %.3f, want %.3f", got, want)
	}
}

func TestEstimateCostUsesCodexBillableEstimate(t *testing.T) {
	event := Event{
		ToolName:        "codex",
		ModelName:       "gpt-5.5",
		InputTokens:     128_415_423,
		OutputTokens:    302_708,
		CacheReadTokens: 120_000_000,
	}

	got := EstimateCost(event)

	if math.Abs(got-65.116) > 0.01 {
		t.Fatalf("codex cost = %.3f, want around 65.116", got)
	}
}

func TestEstimateCostUsesClaudeCacheRates(t *testing.T) {
	event := Event{
		ModelName:        "claude-sonnet-4",
		InputTokens:      1_000_000,
		OutputTokens:     100_000,
		CacheReadTokens:  900_000,
		CacheWriteTokens: 50_000,
		ReasoningTokens:  25_000,
		ToolTokens:       25_000,
	}

	got := EstimateCost(event)
	want := 100_000.0/1_000_000*3.00 +
		900_000.0/1_000_000*0.30 +
		50_000.0/1_000_000*3.75 +
		150_000.0/1_000_000*15.00

	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("cost = %.3f, want %.3f", got, want)
	}
}
