package report

import (
	"strings"
	"testing"
	"time"

	"github.com/why19970628/agentmeter/internal/usage"
)

func TestFilterPeriodTodayUsesLocalDay(t *testing.T) {
	now := time.Date(2026, 5, 23, 15, 0, 0, 0, time.Local)
	events := []usage.Event{
		{ToolName: "codex", TotalTokens: 10, OccurredAt: time.Date(2026, 5, 23, 8, 0, 0, 0, time.Local)},
		{ToolName: "codex", TotalTokens: 20, OccurredAt: time.Date(2026, 5, 22, 8, 0, 0, 0, time.Local)},
	}

	got := FilterPeriod(events, "today", now)

	if len(got) != 1 || got[0].TotalTokens != 10 {
		t.Fatalf("filtered events = %+v, want only today's event", got)
	}
}

func TestRenderTableIncludesTotalsAndCosts(t *testing.T) {
	events := []usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 1000, OutputTokens: 500, TotalTokens: 1500},
		{ToolName: "claude-code", ModelName: "claude-sonnet-4", InputTokens: 2000, OutputTokens: 1000, TotalTokens: 3000},
	}

	got := Render(events, Options{Format: "table"})

	for _, want := range []string{"Source", "Codex", "Claude Code", "Total", "4,500"} {
		if !strings.Contains(got, want) {
			t.Fatalf("table output missing %q:\n%s", want, got)
		}
	}
}

func TestRenderMarkdownIncludesPipeTable(t *testing.T) {
	events := []usage.Event{{ToolName: "gemini", ModelName: "gemini-2.5-pro", TotalTokens: 100}}

	got := Render(events, Options{Format: "markdown"})

	if !strings.Contains(got, "| Source |") || !strings.Contains(got, "Gemini") {
		t.Fatalf("markdown output = %s", got)
	}
}

func TestRenderTableCanGroupByModel(t *testing.T) {
	events := []usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 1000, OutputTokens: 500, TotalTokens: 1500},
		{ToolName: "claude-code", ModelName: "claude-sonnet-4", InputTokens: 2000, OutputTokens: 1000, TotalTokens: 3000},
		{ToolName: "cursor", ModelName: "claude-sonnet-4", InputTokens: 300, OutputTokens: 200, TotalTokens: 500},
	}

	got := Render(events, Options{Format: "table", Group: "model"})

	for _, want := range []string{"Model", "gpt-5", "claude-sonnet-4", "3,500", "Total"} {
		if !strings.Contains(got, want) {
			t.Fatalf("model table output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Codex") {
		t.Fatalf("model grouped output should not use source names:\n%s", got)
	}
}

func TestRenderTableCanGroupByTool(t *testing.T) {
	events := []usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 1000, OutputTokens: 500, TotalTokens: 1500},
		{ToolName: "codex", ModelName: "gpt-5.5", InputTokens: 2000, OutputTokens: 1000, TotalTokens: 3000},
		{ToolName: "cursor", ModelName: "claude-sonnet-4", InputTokens: 300, OutputTokens: 200, TotalTokens: 500},
	}

	got := Render(events, Options{Format: "table", Group: "tool"})

	header := strings.Fields(strings.SplitN(got, "\n", 2)[0])
	if len(header) == 0 || header[0] != "Tool" {
		t.Fatalf("tool grouped table first header = %q, want Tool:\n%s", header, got)
	}
	for _, want := range []string{"Codex", "Cursor", "4,500", "Total"} {
		if !strings.Contains(got, want) {
			t.Fatalf("tool table output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "gpt-5") || strings.Contains(got, "claude-sonnet-4") {
		t.Fatalf("tool grouped output should not use model names:\n%s", got)
	}
}

func TestRenderTableCanGroupByToolAndModel(t *testing.T) {
	events := []usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 1000, OutputTokens: 500, TotalTokens: 1500},
		{ToolName: "codex", ModelName: "gpt-5.5", InputTokens: 2000, OutputTokens: 1000, TotalTokens: 3000},
		{ToolName: "cursor", ModelName: "gpt-5", InputTokens: 300, OutputTokens: 200, TotalTokens: 500},
	}

	got := Render(events, Options{Format: "table", Group: "tool,model"})

	headerLine := strings.SplitN(got, "\n", 2)[0]
	if !strings.HasPrefix(headerLine, "Tool / Model") {
		t.Fatalf("tool+model grouped header = %q, want Tool / Model", headerLine)
	}
	for _, want := range []string{"Codex / gpt-5", "Codex / gpt-5.5", "Cursor / gpt-5", "Total"} {
		if !strings.Contains(got, want) {
			t.Fatalf("tool+model table output missing %q:\n%s", want, got)
		}
	}
}

func TestNormalizeGroupSupportsRepeatedAndCommaSeparatedGroups(t *testing.T) {
	got := NormalizeGroup([]string{"tool", "model"})
	if got != "tool,model" {
		t.Fatalf("NormalizeGroup repeated = %q, want tool,model", got)
	}
	got = NormalizeGroup([]string{"tool,model"})
	if got != "tool,model" {
		t.Fatalf("NormalizeGroup comma separated = %q, want tool,model", got)
	}
	got = NormalizeGroup([]string{"model", "tool", "model"})
	if got != "tool,model" {
		t.Fatalf("NormalizeGroup should dedupe and keep canonical order = %q", got)
	}
}

func TestRenderTableIncludesExtendedTokenColumns(t *testing.T) {
	events := []usage.Event{
		{
			ToolName:         "claude-code",
			ModelName:        "claude-sonnet-4",
			InputTokens:      100,
			OutputTokens:     40,
			CacheReadTokens:  20,
			CacheWriteTokens: 10,
			ReasoningTokens:  7,
			ToolTokens:       3,
			TotalTokens:      160,
		},
	}

	got := Render(events, Options{Format: "table", Group: "model"})

	for _, want := range []string{"Cache Read", "Cache Write", "Reasoning", "Tool Tokens", "20", "10", "7"} {
		if !strings.Contains(got, want) {
			t.Fatalf("extended token table output missing %q:\n%s", want, got)
		}
	}
}

func TestRenderTableSupportsChineseHeaders(t *testing.T) {
	events := []usage.Event{{ToolName: "codex", ModelName: "gpt-5", InputTokens: 100, TotalTokens: 100}}

	got := Render(events, Options{Format: "table", Group: "model", Lang: "zh-CN"})

	for _, want := range []string{"模型", "输入 Tokens", "总 Tokens", "预估费用"} {
		if !strings.Contains(got, want) {
			t.Fatalf("chinese table output missing %q:\n%s", want, got)
		}
	}
}
