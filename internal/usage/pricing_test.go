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

func TestEstimateCostUsesModelSpecificCachedInputRate(t *testing.T) {
	event := Event{
		ToolName:        "codex",
		ModelName:       "gpt-5.5",
		InputTokens:     10_000_000,
		OutputTokens:    100_000,
		CacheReadTokens: 6_000_000,
	}

	got := EstimateCost(event)
	want := 4_000_000.0/1_000_000*5.00 +
		6_000_000.0/1_000_000*0.50 +
		100_000.0/1_000_000*30.00

	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("cost = %.3f, want %.3f", got, want)
	}
}

func TestEstimateCostMatchesCodexHistoricalBillingSamples(t *testing.T) {
	tests := []struct {
		model         string
		billableInput int64
		output        int64
		cacheRead     int64
		want          float64
	}{
		{"gpt-5.4", 51_634, 5_367, 566_272, 0.3512},
		{"gpt-5.4-mini", 9_969, 195, 7_168, 0.0089},
		{"gpt-5.5", 215_692, 22_262, 4_043_904, 3.7683},
		{"gpt-5.4", 330_738, 19_309, 2_324_864, 1.6977},
		{"gpt-5.4-mini", 19_928, 284, 14_336, 0.0173},
		{"gpt-5.5", 2_749_814, 207_766, 79_391_360, 59.6777},
		{"gpt-5.4", 100_734, 9_575, 1_280_640, 0.7156},
		{"gpt-5.4-mini", 32_223, 388, 24_576, 0.0278},
		{"gpt-5.5", 937_193, 83_081, 36_115_072, 25.2359},
		{"gpt-5.4", 100_442, 10_714, 1_088_512, 0.6839},
		{"gpt-5.4-mini", 14_233, 204, 9_216, 0.0123},
		{"gpt-5.5", 34_956, 1_593, 66_432, 0.2558},
		{"gpt-5.3-codex", 18_824, 381, 47_616, 0.0466},
		{"gpt-5.4", 198_556, 12_230, 1_045_120, 0.9411},
		{"gpt-5.4-mini", 18_810, 199, 11_264, 0.0158},
		{"gpt-5.5", 2_032_717, 127_401, 54_552_960, 41.2621},
		{"gpt-5.4", 284_545, 14_659, 1_144_704, 1.2174},
		{"gpt-5.4-mini", 24_994, 361, 22_528, 0.0221},
		{"gpt-5.5", 2_182_820, 230_835, 88_184_832, 61.9316},
		{"gpt-5.4", 134_846, 13_484, 1_744_640, 0.9755},
		{"gpt-5.4-mini", 14_797, 169, 9_216, 0.0125},
		{"gpt-5.5", 400_935, 71_857, 32_254_976, 20.2879},
		{"gpt-5.4", 95_794, 7_404, 697_344, 0.5249},
		{"gpt-5.4-mini", 20_124, 325, 14_336, 0.0176},
		{"gpt-5.5", 2_345_570, 117_386, 50_670_336, 40.5846},
		{"gpt-5.4", 98_575, 10_328, 841_984, 0.6119},
		{"gpt-5.4-mini", 14_228, 150, 9_216, 0.0120},
		{"gpt-5.5", 2_264_701, 148_695, 63_461_248, 47.5150},
		{"gpt-5.4", 114_170, 9_698, 723_072, 0.6117},
		{"gpt-5.4-mini", 18_969, 276, 14_336, 0.0165},
		{"gpt-5.5", 738_947, 90_201, 26_602_240, 19.7019},
		{"gpt-5.4", 178_285, 13_470, 635_776, 0.8067},
		{"gpt-5.4-mini", 26_048, 301, 16_896, 0.0222},
		{"gpt-5.5", 838_015, 97_913, 26_532_736, 20.3938},
	}

	for _, tt := range tests {
		event := Event{
			ToolName:        "codex",
			ModelName:       tt.model,
			InputTokens:     tt.billableInput + tt.cacheRead,
			OutputTokens:    tt.output,
			CacheReadTokens: tt.cacheRead,
		}

		got := EstimateCost(event)
		if math.Abs(got-tt.want) > 0.0006 {
			t.Fatalf("%s cost = %.4f, want %.4f", tt.model, got, tt.want)
		}
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
