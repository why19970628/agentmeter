package usage

import "strings"

const codexBillableEstimateMultiplier = 0.10

type ModelPrice struct {
	InputPerMTokens       float64
	CachedInputPerMTokens float64
	CacheWritePerMTokens  float64
	OutputPerMTokens      float64
}

type CostBreakdown struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cache_read"`
	CacheWrite float64 `json:"cache_write"`
	Reasoning  float64 `json:"reasoning"`
	Tool       float64 `json:"tool"`
	Total      float64 `json:"total"`
	Note       string  `json:"note"`
}

type modelPriceEntry struct {
	Match string
	Price ModelPrice
}

var builtInPrices = []modelPriceEntry{
	{Match: "gpt-5.5", Price: ModelPrice{InputPerMTokens: 5.00, CachedInputPerMTokens: 0.50, OutputPerMTokens: 30.00}},
	{Match: "gpt-5", Price: ModelPrice{InputPerMTokens: 1.25, CachedInputPerMTokens: 0.125, OutputPerMTokens: 10.00}},
	{Match: "gpt-4.1", Price: ModelPrice{InputPerMTokens: 2.00, CachedInputPerMTokens: 0.50, OutputPerMTokens: 8.00}},
	{Match: "gpt-4o", Price: ModelPrice{InputPerMTokens: 2.50, CachedInputPerMTokens: 1.25, OutputPerMTokens: 10.00}},
	{Match: "claude-sonnet-4", Price: ModelPrice{InputPerMTokens: 3.00, CachedInputPerMTokens: 0.30, CacheWritePerMTokens: 3.75, OutputPerMTokens: 15.00}},
	{Match: "claude-3.5", Price: ModelPrice{InputPerMTokens: 3.00, CachedInputPerMTokens: 0.30, CacheWritePerMTokens: 3.75, OutputPerMTokens: 15.00}},
	{Match: "gemini-2.5-pro", Price: ModelPrice{InputPerMTokens: 1.25, CachedInputPerMTokens: 0.31, OutputPerMTokens: 10.00}},
	{Match: "gemini-1.5-pro", Price: ModelPrice{InputPerMTokens: 1.25, CachedInputPerMTokens: 0.31, OutputPerMTokens: 5.00}},
}

func EstimateCost(event Event) float64 {
	return EstimateCostBreakdown(event).Total
}

func EstimateCostBreakdown(event Event) CostBreakdown {
	if event.CostAmount > 0 {
		return CostBreakdown{
			Total: event.CostAmount,
			Note:  "Cost came from the source log, so column-level split is unavailable.",
		}
	}
	price, ok := priceForModel(event.ModelName)
	if !ok {
		return CostBreakdown{Note: "No built-in price rule for this model."}
	}
	if isCodexTool(event.ToolName) {
		inputSideTokens := maxInt64(event.InputTokens, event.CacheReadTokens)
		inputCost := perMillion(inputSideTokens, price.InputPerMTokens) * codexBillableEstimateMultiplier
		outputTokens := event.OutputTokens + event.ReasoningTokens + event.ToolTokens
		breakdownInput := inputCost
		breakdownCacheRead := 0.0
		if event.CacheReadTokens > event.InputTokens {
			breakdownInput = perMillion(event.InputTokens, price.InputPerMTokens) * codexBillableEstimateMultiplier
			breakdownCacheRead = perMillion(event.CacheReadTokens, price.InputPerMTokens) * codexBillableEstimateMultiplier
		}
		breakdown := CostBreakdown{
			Input:     breakdownInput,
			CacheRead: breakdownCacheRead,
			Output:    perMillion(event.OutputTokens, price.OutputPerMTokens) * codexBillableEstimateMultiplier,
			Reasoning: perMillion(event.ReasoningTokens, price.OutputPerMTokens) * codexBillableEstimateMultiplier,
			Tool:      perMillion(event.ToolTokens, price.OutputPerMTokens) * codexBillableEstimateMultiplier,
			Note:      "Codex estimate uses the larger of input tokens and cache read as input-side usage, then applies the local billable multiplier.",
		}
		breakdown.Total = inputCost + breakdown.Output + breakdown.Reasoning + breakdown.Tool
		if outputTokens == 0 {
			breakdown.Output = 0
		}
		return breakdown
	}
	billedInput := maxInt64(0, event.InputTokens-event.CacheReadTokens)
	breakdown := CostBreakdown{
		Input:      perMillion(billedInput, price.InputPerMTokens),
		CacheRead:  perMillion(event.CacheReadTokens, price.CachedInputPerMTokens),
		CacheWrite: perMillion(event.CacheWriteTokens, price.CacheWritePerMTokens),
		Output:     perMillion(event.OutputTokens, price.OutputPerMTokens),
		Reasoning:  perMillion(event.ReasoningTokens, price.OutputPerMTokens),
		Tool:       perMillion(event.ToolTokens, price.OutputPerMTokens),
	}
	breakdown.Total = breakdown.Input + breakdown.CacheRead + breakdown.CacheWrite + breakdown.Output + breakdown.Reasoning + breakdown.Tool
	return breakdown
}

func isCodexTool(tool string) bool {
	return strings.EqualFold(strings.TrimSpace(tool), "codex")
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
