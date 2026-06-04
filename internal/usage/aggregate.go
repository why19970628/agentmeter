package usage

import (
	"fmt"
	"sort"
	"time"
)

func BuildDashboard(events []Event, grain Grain) Dashboard {
	if grain == "" {
		grain = GrainDay
	}

	summary := Summary{}
	models := map[string]struct{}{}
	tools := map[string]struct{}{}
	buckets := map[string]*Bucket{}
	toolRank := map[string]*RankingItem{}
	modelRank := map[string]*RankingItem{}
	projectRank := map[string]*RankingItem{}

	for _, event := range events {
		total := normalizedTotal(event)

		summary.InputTokens += BillableInputTokens(event)
		summary.OutputTokens += event.OutputTokens
		summary.CacheReadTokens += event.CacheReadTokens
		summary.CacheWriteTokens += event.CacheWriteTokens
		summary.ReasoningTokens += event.ReasoningTokens
		summary.ToolTokens += event.ToolTokens
		summary.TotalTokens += total
		summary.RequestCount++
		summary.CostAmount += EstimateCost(event)

		if event.ModelName != "" {
			models[event.ModelName] = struct{}{}
		}
		if event.ToolName != "" {
			tools[event.ToolName] = struct{}{}
		}

		addBucket(buckets, bucketKey(event.OccurredAt, grain), event, total)
		addRanking(toolRank, emptyAs(event.ToolName, "unknown"), event, total)
		addRanking(modelRank, emptyAs(event.ModelName, "unknown"), event, total)
		addRanking(projectRank, emptyAs(event.ProjectPath, "unknown"), event, total)
	}

	summary.ModelCount = len(models)
	summary.ToolCount = len(tools)

	return Dashboard{
		Summary:        summary,
		Trend:          sortedBuckets(buckets),
		ToolRanking:    sortedRanking(toolRank),
		ModelRanking:   sortedRanking(modelRank),
		ProjectRanking: sortedRanking(projectRank),
	}
}

func addBucket(items map[string]*Bucket, key string, event Event, total int64) {
	item, ok := items[key]
	if !ok {
		item = &Bucket{Key: key}
		items[key] = item
	}
	item.InputTokens += BillableInputTokens(event)
	item.OutputTokens += event.OutputTokens
	item.CacheReadTokens += event.CacheReadTokens
	item.CacheWriteTokens += event.CacheWriteTokens
	item.ReasoningTokens += event.ReasoningTokens
	item.ToolTokens += event.ToolTokens
	item.TotalTokens += total
	item.RequestCount++
	item.CostAmount += EstimateCost(event)
}

func addRanking(items map[string]*RankingItem, name string, event Event, total int64) {
	item, ok := items[name]
	if !ok {
		item = &RankingItem{Name: name}
		items[name] = item
	}
	item.InputTokens += BillableInputTokens(event)
	item.OutputTokens += event.OutputTokens
	item.CacheReadTokens += event.CacheReadTokens
	item.CacheWriteTokens += event.CacheWriteTokens
	item.ReasoningTokens += event.ReasoningTokens
	item.ToolTokens += event.ToolTokens
	item.TotalTokens += total
	item.RequestCount++
	item.CostAmount += EstimateCost(event)
}

func sortedBuckets(items map[string]*Bucket) []Bucket {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]Bucket, 0, len(keys))
	for _, key := range keys {
		out = append(out, *items[key])
	}
	return out
}

func sortedRanking(items map[string]*RankingItem) []RankingItem {
	out := make([]RankingItem, 0, len(items))
	for _, item := range items {
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TotalTokens == out[j].TotalTokens {
			return out[i].Name < out[j].Name
		}
		return out[i].TotalTokens > out[j].TotalTokens
	})
	return out
}

func bucketKey(t time.Time, grain Grain) string {
	local := t.In(time.Local)
	switch grain {
	case GrainWeek:
		year, week := local.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case GrainMonth:
		return local.Format("2006-01")
	default:
		return local.Format("2006-01-02")
	}
}

func emptyAs(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func normalizedTotal(event Event) int64 {
	if event.TotalTokens > 0 {
		return event.TotalTokens
	}
	return event.InputTokens + event.OutputTokens + event.CacheWriteTokens + event.ReasoningTokens + event.ToolTokens
}
