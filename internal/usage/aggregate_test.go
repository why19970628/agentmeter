package usage

import (
	"testing"
	"time"
)

func TestBuildDashboardAggregatesByDayAndRanksTools(t *testing.T) {
	events := []Event{
		{
			ToolName: "codex", ModelName: "gpt-5", ProjectPath: "/work/a",
			InputTokens: 100, OutputTokens: 40, TotalTokens: 140, CostAmount: 0.12,
			OccurredAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ToolName: "claude-code", ModelName: "claude-sonnet", ProjectPath: "/work/b",
			InputTokens: 200, OutputTokens: 80, TotalTokens: 280, CostAmount: 0.30,
			OccurredAt: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ToolName: "codex", ModelName: "gpt-5", ProjectPath: "/work/a",
			InputTokens: 50, OutputTokens: 10, TotalTokens: 60, CostAmount: 0.05,
			OccurredAt: time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		},
	}

	got := BuildDashboard(events, GrainDay)

	if got.Summary.TotalTokens != 480 {
		t.Fatalf("summary total tokens = %d, want 480", got.Summary.TotalTokens)
	}
	if got.Summary.RequestCount != 3 {
		t.Fatalf("summary request count = %d, want 3", got.Summary.RequestCount)
	}
	if got.Summary.ModelCount != 2 {
		t.Fatalf("summary model count = %d, want 2", got.Summary.ModelCount)
	}
	if len(got.Trend) != 2 {
		t.Fatalf("trend buckets = %d, want 2", len(got.Trend))
	}
	if got.Trend[0].Key != "2026-05-01" || got.Trend[0].TotalTokens != 420 {
		t.Fatalf("first bucket = %+v, want key 2026-05-01 total 420", got.Trend[0])
	}
	if got.ToolRanking[0].Name != "claude-code" || got.ToolRanking[0].TotalTokens != 280 {
		t.Fatalf("top tool = %+v, want claude-code total 280", got.ToolRanking[0])
	}
	if got.ProjectRanking[0].Name != "/work/b" || got.ProjectRanking[0].TotalTokens != 280 {
		t.Fatalf("top project = %+v, want /work/b total 280", got.ProjectRanking[0])
	}
	if got.ProjectRanking[1].Name != "/work/a" || got.ProjectRanking[1].TotalTokens != 200 {
		t.Fatalf("second project = %+v, want /work/a total 200", got.ProjectRanking[1])
	}
}
