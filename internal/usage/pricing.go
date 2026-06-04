package usage

import "strings"

type ModelPrice struct {
	InputPerMTokens       float64
	CachedInputPerMTokens float64
	CacheWritePerMTokens  float64
	OutputPerMTokens      float64
}

type modelPriceEntry struct {
	Match string
	Price ModelPrice
}

var builtInPrices = []modelPriceEntry{
	{Match: "gpt-5.5", Price: ModelPrice{InputPerMTokens: 5.00, CachedInputPerMTokens: 0.50, OutputPerMTokens: 30.00}},
	{Match: "gpt-5.4-mini", Price: ModelPrice{InputPerMTokens: 0.744226, CachedInputPerMTokens: 0.086353, OutputPerMTokens: 4.315918}},
	{Match: "gpt-5.4", Price: ModelPrice{InputPerMTokens: 2.50, CachedInputPerMTokens: 0.25, OutputPerMTokens: 15.00}},
	{Match: "gpt-5.3-codex", Price: ModelPrice{InputPerMTokens: 2.00, CachedInputPerMTokens: 0.10, OutputPerMTokens: 10.00}},
	{Match: "gpt-5", Price: ModelPrice{InputPerMTokens: 1.25, CachedInputPerMTokens: 0.125, OutputPerMTokens: 10.00}},
	{Match: "gpt-4.1", Price: ModelPrice{InputPerMTokens: 2.00, CachedInputPerMTokens: 0.50, OutputPerMTokens: 8.00}},
	{Match: "gpt-4o", Price: ModelPrice{InputPerMTokens: 2.50, CachedInputPerMTokens: 1.25, OutputPerMTokens: 10.00}},
	{Match: "claude-sonnet-4", Price: ModelPrice{InputPerMTokens: 3.00, CachedInputPerMTokens: 0.30, CacheWritePerMTokens: 3.75, OutputPerMTokens: 15.00}},
	{Match: "claude-3.5", Price: ModelPrice{InputPerMTokens: 3.00, CachedInputPerMTokens: 0.30, CacheWritePerMTokens: 3.75, OutputPerMTokens: 15.00}},
	{Match: "gemini-2.5-pro", Price: ModelPrice{InputPerMTokens: 1.25, CachedInputPerMTokens: 0.31, OutputPerMTokens: 10.00}},
	{Match: "gemini-1.5-pro", Price: ModelPrice{InputPerMTokens: 1.25, CachedInputPerMTokens: 0.31, OutputPerMTokens: 5.00}},
}

func EstimateCost(event Event) float64 {
	if event.CostAmount > 0 {
		return event.CostAmount
	}
	price, ok := priceForModel(event.ModelName)
	if !ok {
		return 0
	}
	billedInput := BillableInputTokens(event)
	billedOutput := event.OutputTokens + event.ReasoningTokens + event.ToolTokens
	base := perMillion(billedInput, price.InputPerMTokens) +
		perMillion(event.CacheReadTokens, price.CachedInputPerMTokens) +
		perMillion(event.CacheWriteTokens, price.CacheWritePerMTokens) +
		perMillion(billedOutput, price.OutputPerMTokens)
	return base
}

func BillableInputTokens(event Event) int64 {
	return maxInt64(0, event.InputTokens-event.CacheReadTokens)
}

func priceForModel(model string) (ModelPrice, bool) {
	model = strings.ToLower(model)
	for _, item := range builtInPrices {
		if strings.Contains(model, item.Match) {
			return item.Price, true
		}
	}
	return ModelPrice{}, false
}

func perMillion(tokens int64, rate float64) float64 {
	return float64(tokens) / 1_000_000 * rate
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
