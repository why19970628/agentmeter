package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"agentmeter/internal/usage"
)

type Options struct {
	Format string
	Group  string
	Lang   string
}

type Row struct {
	Source           string  `json:"source"`
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	CacheReadTokens  int64   `json:"cache_read_tokens"`
	CacheWriteTokens int64   `json:"cache_write_tokens"`
	ReasoningTokens  int64   `json:"reasoning_tokens"`
	ToolTokens       int64   `json:"tool_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	EstUSD           float64 `json:"est_usd"`
}

func FilterPeriod(events []usage.Event, period string, now time.Time) []usage.Event {
	if period == "" || period == "all" {
		return events
	}
	start := periodStart(period, now)
	if start.IsZero() {
		return events
	}
	out := make([]usage.Event, 0, len(events))
	for _, event := range events {
		localTime := event.OccurredAt.In(time.Local)
		if !localTime.Before(start) && !localTime.After(now) {
			out = append(out, event)
		}
	}
	return out
}

func Render(events []usage.Event, opts Options) string {
	rows := BuildRows(events, opts.Group)
	label := groupLabel(opts.Group)
	if opts.Lang == "zh-CN" {
		label = groupLabelZH(opts.Group)
		for i := range rows {
			if rows[i].Source == "Total" {
				rows[i].Source = "总计"
			}
		}
	}
	switch opts.Format {
	case "json":
		data, _ := json.MarshalIndent(rows, "", "  ")
		return string(data) + "\n"
	case "markdown":
		return renderMarkdownWithLang(rows, label, opts.Lang)
	default:
		return renderTableWithLang(rows, label, opts.Lang)
	}
}

func BuildRows(events []usage.Event, group string) []Row {
	byGroup := map[string]*Row{}
	for _, event := range events {
		name := groupName(event, group)
		row, ok := byGroup[name]
		if !ok {
			row = &Row{Source: name}
			byGroup[name] = row
		}
		row.InputTokens += event.InputTokens
		row.OutputTokens += event.OutputTokens
		row.CacheReadTokens += event.CacheReadTokens
		row.CacheWriteTokens += event.CacheWriteTokens
		row.ReasoningTokens += event.ReasoningTokens
		row.ToolTokens += event.ToolTokens
		row.TotalTokens += normalizedTotal(event)
		row.EstUSD += usage.EstimateCost(event)
	}

	rows := make([]Row, 0, len(byGroup))
	for _, row := range byGroup {
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].TotalTokens == rows[j].TotalTokens {
			return rows[i].Source < rows[j].Source
		}
		return rows[i].TotalTokens > rows[j].TotalTokens
	})

	total := Row{Source: "Total"}
	for _, row := range rows {
		total.InputTokens += row.InputTokens
		total.OutputTokens += row.OutputTokens
		total.CacheReadTokens += row.CacheReadTokens
		total.CacheWriteTokens += row.CacheWriteTokens
		total.ReasoningTokens += row.ReasoningTokens
		total.ToolTokens += row.ToolTokens
		total.TotalTokens += row.TotalTokens
		total.EstUSD += row.EstUSD
	}
	rows = append(rows, total)
	return rows
}

func renderTable(rows []Row) string {
	return renderTableWithLabel(rows, "Source")
}

func renderTableWithLabel(rows []Row, label string) string {
	return renderTableWithLang(rows, label, "")
}

func renderTableWithLang(rows []Row, label string, lang string) string {
	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	headers := []string{label, "Input Tokens", "Output Tokens", "Cache Read", "Cache Write", "Reasoning", "Tool Tokens", "Total Tokens", "Est. USD"}
	if lang == "zh-CN" {
		headers = []string{label, "输入 Tokens", "输出 Tokens", "缓存读取", "缓存写入", "推理 Tokens", "工具 Tokens", "总 Tokens", "预估费用(USD)"}
	}
	fmt.Fprintf(writer, "%s\n", strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t$%.3f\n",
			row.Source,
			formatInt(row.InputTokens),
			formatInt(row.OutputTokens),
			formatInt(row.CacheReadTokens),
			formatInt(row.CacheWriteTokens),
			formatInt(row.ReasoningTokens),
			formatInt(row.ToolTokens),
			formatInt(row.TotalTokens),
			row.EstUSD,
		)
	}
	_ = writer.Flush()
	return buf.String()
}

func renderMarkdown(rows []Row) string {
	return renderMarkdownWithLabel(rows, "Source")
}

func renderMarkdownWithLabel(rows []Row, label string) string {
	return renderMarkdownWithLang(rows, label, "")
}

func renderMarkdownWithLang(rows []Row, label string, lang string) string {
	var buf bytes.Buffer
	headers := []string{label, "Input Tokens", "Output Tokens", "Cache Read", "Cache Write", "Reasoning", "Tool Tokens", "Total Tokens", "Est. USD"}
	if lang == "zh-CN" {
		headers = []string{label, "输入 Tokens", "输出 Tokens", "缓存读取", "缓存写入", "推理 Tokens", "工具 Tokens", "总 Tokens", "预估费用(USD)"}
	}
	fmt.Fprintf(&buf, "| %s |\n", strings.Join(headers, " | "))
	buf.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, row := range rows {
		fmt.Fprintf(&buf, "| %s | %s | %s | %s | %s | %s | %s | %s | $%.3f |\n",
			row.Source,
			formatInt(row.InputTokens),
			formatInt(row.OutputTokens),
			formatInt(row.CacheReadTokens),
			formatInt(row.CacheWriteTokens),
			formatInt(row.ReasoningTokens),
			formatInt(row.ToolTokens),
			formatInt(row.TotalTokens),
			row.EstUSD,
		)
	}
	return buf.String()
}

func normalizedTotal(event usage.Event) int64 {
	if event.TotalTokens > 0 {
		return event.TotalTokens
	}
	return event.InputTokens + event.OutputTokens + event.CacheWriteTokens + event.ReasoningTokens + event.ToolTokens
}

func groupLabel(group string) string {
	if group == "model" {
		return "Model"
	}
	return "Source"
}

func groupLabelZH(group string) string {
	if group == "model" {
		return "模型"
	}
	return "来源"
}

func groupName(event usage.Event, group string) string {
	if group == "model" {
		if event.ModelName == "" {
			return "unknown"
		}
		return event.ModelName
	}
	return displaySource(event.ToolName)
}

func periodStart(period string, now time.Time) time.Time {
	localNow := now.In(time.Local)
	switch period {
	case "today":
		return time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.Local)
	case "week":
		weekday := int(localNow.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		day := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.Local)
		return day.AddDate(0, 0, -(weekday - 1))
	case "month":
		return time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, time.Local)
	default:
		return time.Time{}
	}
}

func displaySource(source string) string {
	switch strings.ToLower(source) {
	case "codex":
		return "Codex"
	case "claude", "claude-code":
		return "Claude Code"
	case "gemini", "gemini-cli":
		return "Gemini CLI"
	case "cursor":
		return "Cursor"
	case "":
		return "Unknown"
	default:
		return source
	}
}

func formatInt(value int64) string {
	raw := fmt.Sprintf("%d", value)
	if len(raw) <= 3 {
		return raw
	}
	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)
	return strings.Join(parts, ",")
}
