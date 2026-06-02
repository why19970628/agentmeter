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

func TestEstimateCostMatchesBreakdownTotal(t *testing.T) {
	event := Event{
		ToolName:        "codex",
		ModelName:       "gpt-5.5",
		InputTokens:     173_670_501,
		OutputTokens:    382_602,
		CacheReadTokens: 168_433_664,
	}

	cost := EstimateCost(event)
	breakdown := EstimateCostBreakdown(event)

	if math.Abs(cost-breakdown.Total) > 0.0001 {
		t.Fatalf("cost = %.3f, breakdown total = %.3f", cost, breakdown.Total)
	}
	if math.Abs(cost-87.983) > 0.01 {
		t.Fatalf("codex cost = %.3f, want today's known estimate around 87.983", cost)
	}
}

func TestEstimateCostUsesCodexCacheReadAsInputSideUsage(t *testing.T) {
	event := Event{
		ToolName:        "codex",
		ModelName:       "gpt-5.5",
		InputTokens:     113_000,
		OutputTokens:    391_300,
		CacheReadTokens: 171_100_000,
		TotalTokens:     176_800_000,
	}

	got := EstimateCost(event)

	if math.Abs(got-86.723) > 0.01 {
		t.Fatalf("codex cost = %.3f, want cache read used as input-side estimate around 86.723", got)
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
