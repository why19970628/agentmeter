package web

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"agentmeter/internal/usage"
)

type Server struct {
	events    []usage.Event
	templates *template.Template
	staticDir string
}

type PageData struct {
	GeneratedAt   string
	Dashboard     usage.Dashboard
	DailyRequests []usage.Bucket
	ModelRows     []UsageRow
	Grain         usage.Grain
	Lang          Lang
	Unresolved    string
}

type UsageRow struct {
	Period           string
	Tool             string
	Name             string
	InputTokens      int64
	OutputTokens     int64
	CacheReadTokens  int64
	CacheWriteTokens int64
	ReasoningTokens  int64
	ToolTokens       int64
	TotalTokens      int64
	EstUSD           float64
}

func NewServer(events []usage.Event, templateDir string, staticDir string) (*Server, error) {
	funcs := template.FuncMap{
		"json":    templateJSON,
		"format":  formatInt,
		"compact": formatCompact,
		"t":       func(data PageData, key string) string { return translate(data.Lang, key) },
	}
	tmpl, err := template.New("layout.html").Funcs(funcs).ParseFiles(
		filepath.Join(templateDir, "layout.html"),
		filepath.Join(templateDir, "index.html"),
	)
	if err != nil {
		return nil, err
	}
	return &Server{events: events, templates: tmpl, staticDir: staticDir}, nil
}

func templateJSON(value any) template.JS {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(value); err != nil {
		return "null"
	}
	return template.JS(buf.String())
}

func formatInt(value int64) string {
	raw := strconv.FormatInt(value, 10)
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

func formatCompact(value int64) string {
	abs := value
	if abs < 0 {
		abs = -abs
	}
	type unit struct {
		value  int64
		suffix string
	}
	for _, u := range []unit{
		{1_000_000_000_000, "T"},
		{1_000_000_000, "B"},
		{1_000_000, "M"},
		{1_000, "K"},
	} {
		if abs >= u.value {
			return fmt.Sprintf("%.1f%s", float64(value)/float64(u.value), u.suffix)
		}
	}
	return formatInt(value)
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.index)
	mux.HandleFunc("/api/dashboard", s.dashboard)
	mux.HandleFunc("/export.csv", s.exportCSV)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))
	return mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	grain := parseGrain(r)
	lang := parseLang(r.URL.Query().Get("lang"))
	data := PageData{
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
		Dashboard:     usage.BuildDashboard(s.events, grain),
		DailyRequests: usage.BuildDashboard(s.events, usage.GrainDay).Trend,
		Grain:         grain,
		Lang:          lang,
		Unresolved:    translate(lang, "unresolved"),
	}
	data.ModelRows = modelRows(s.events, data.Unresolved)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func modelRows(events []usage.Event, unresolved string) []UsageRow {
	byModel := map[string]*UsageRow{}
	for _, event := range events {
		period := event.OccurredAt.Format("2006-01-02")
		model := event.ModelName
		if model == "" {
			model = unresolved
		}
		tool := displayTool(event.ToolName)
		key := period + "\x00" + tool + "\x00" + model
		row, ok := byModel[key]
		if !ok {
			row = &UsageRow{Period: period, Tool: tool, Name: model}
			byModel[key] = row
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

	rows := make([]UsageRow, 0, len(byModel)+1)
	total := UsageRow{Period: "Total", Tool: "Total", Name: "Total"}
	for _, row := range byModel {
		rows = append(rows, *row)
		total.InputTokens += row.InputTokens
		total.OutputTokens += row.OutputTokens
		total.CacheReadTokens += row.CacheReadTokens
		total.CacheWriteTokens += row.CacheWriteTokens
		total.ReasoningTokens += row.ReasoningTokens
		total.ToolTokens += row.ToolTokens
		total.TotalTokens += row.TotalTokens
		total.EstUSD += row.EstUSD
	}
	sortUsageRows(rows)
	if len(rows) > 0 {
		rows = append(rows, total)
	}
	return rows
}

func normalizedTotal(event usage.Event) int64 {
	if event.TotalTokens > 0 {
		return event.TotalTokens
	}
	return event.InputTokens + event.OutputTokens + event.CacheWriteTokens + event.ReasoningTokens + event.ToolTokens
}

func sortUsageRows(rows []UsageRow) {
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].Period > rows[i].Period ||
				rows[j].Period == rows[i].Period && rows[j].TotalTokens > rows[i].TotalTokens ||
				rows[j].Period == rows[i].Period && rows[j].TotalTokens == rows[i].TotalTokens && rows[j].Tool < rows[i].Tool ||
				rows[j].Period == rows[i].Period && rows[j].TotalTokens == rows[i].TotalTokens && rows[j].Tool == rows[i].Tool && rows[j].Name < rows[i].Name {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}
}

func displayTool(tool string) string {
	switch strings.ToLower(tool) {
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
		return tool
	}
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	grain := parseGrain(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(usage.BuildDashboard(s.events, grain)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) exportCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="agentmeter-usage.csv"`)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{
		"occurred_at", "tool", "model", "session", "project",
		"input_tokens", "output_tokens", "cache_read_tokens", "cache_write_tokens",
		"reasoning_tokens", "tool_tokens", "total_tokens", "cost", "source_file",
	})
	for _, event := range s.events {
		_ = writer.Write([]string{
			event.OccurredAt.Format(time.RFC3339),
			event.ToolName,
			event.ModelName,
			event.SessionID,
			event.ProjectPath,
			strconv.FormatInt(event.InputTokens, 10),
			strconv.FormatInt(event.OutputTokens, 10),
			strconv.FormatInt(event.CacheReadTokens, 10),
			strconv.FormatInt(event.CacheWriteTokens, 10),
			strconv.FormatInt(event.ReasoningTokens, 10),
			strconv.FormatInt(event.ToolTokens, 10),
			strconv.FormatInt(event.TotalTokens, 10),
			fmt.Sprintf("%.6f", event.CostAmount),
			event.SourceFile,
		})
	}
}

func parseGrain(r *http.Request) usage.Grain {
	switch usage.Grain(r.URL.Query().Get("grain")) {
	case usage.GrainWeek:
		return usage.GrainWeek
	case usage.GrainMonth:
		return usage.GrainMonth
	default:
		return usage.GrainDay
	}
}
