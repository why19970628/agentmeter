package usage

import "strings"

type ModelPrice struct {
	InputPerMTokens  float64
	OutputPerMTokens float64
}

var builtInPrices = map[string]ModelPrice{
	"gpt-5":           {InputPerMTokens: 1.25, OutputPerMTokens: 10.00},
	"gpt-4.1":         {InputPerMTokens: 2.00, OutputPerMTokens: 8.00},
	"gpt-4o":          {InputPerMTokens: 2.50, OutputPerMTokens: 10.00},
	"claude-sonnet-4": {InputPerMTokens: 3.00, OutputPerMTokens: 15.00},
	"claude-3.5":      {InputPerMTokens: 3.00, OutputPerMTokens: 15.00},
	"gemini-2.5-pro":  {InputPerMTokens: 1.25, OutputPerMTokens: 10.00},
	"gemini-1.5-pro":  {InputPerMTokens: 1.25, OutputPerMTokens: 5.00},
}

func EstimateCost(event Event) float64 {
	if event.CostAmount > 0 {
		return event.CostAmount
	}
	price, ok := priceForModel(event.ModelName)
	if !ok {
		return 0
	}
	return float64(event.InputTokens)/1_000_000*price.InputPerMTokens +
		float64(event.OutputTokens)/1_000_000*price.OutputPerMTokens
}

func priceForModel(model string) (ModelPrice, bool) {
	model = strings.ToLower(model)
	for key, price := range builtInPrices {
		if strings.Contains(model, key) {
			return price, true
		}
	}
	return ModelPrice{}, false
}
