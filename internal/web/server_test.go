package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/why19970628/agentmeter/internal/usage"
)

func TestMain(m *testing.M) {
	old := time.Local
	time.Local = time.UTC
	code := m.Run()
	time.Local = old
	os.Exit(code)
}

func TestServerRendersIndexAndDashboardAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{
			ToolName:         "codex",
			ModelName:        "gpt-5",
			InputTokens:      100,
			OutputTokens:     10,
			CacheReadTokens:  20,
			CacheWriteTokens: 5,
			ReasoningTokens:  4,
			ToolTokens:       1,
			TotalTokens:      120,
			OccurredAt:       time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	indexReq := httptest.NewRequest(http.MethodGet, "/", nil)
	indexResp := httptest.NewRecorder()
	server.Routes().ServeHTTP(indexResp, indexReq)
	if indexResp.Code != http.StatusOK {
		t.Fatalf("index status = %d, want 200", indexResp.Code)
	}
	if !strings.Contains(indexResp.Body.String(), "AgentMeter") {
		t.Fatalf("index body does not contain AgentMeter")
	}
	if !strings.Contains(indexResp.Body.String(), "Daily API Requests") {
		t.Fatalf("index body does not contain daily request chart title")
	}
	if !strings.Contains(indexResp.Body.String(), "request_count") {
		t.Fatalf("index body does not expose daily request count")
	}
	for _, want := range []string{"Model Usage Details", "Model", "Input Tokens", "Output Tokens", "Total Tokens", "Est. USD", "gpt-5"} {
		if !strings.Contains(indexResp.Body.String(), want) {
			t.Fatalf("index body does not contain model usage table fragment %q", want)
		}
	}
	for _, want := range []string{"Cost estimate", "Codex", "0.10", "Local estimate"} {
		if !strings.Contains(indexResp.Body.String(), want) {
			t.Fatalf("index body does not contain pricing formula fragment %q", want)
		}
	}
	if !strings.Contains(indexResp.Body.String(), `class="estimate-tip"`) || strings.Contains(indexResp.Body.String(), `class="pricing-note"`) || strings.Contains(indexResp.Body.String(), `class="pricing-popover"`) {
		t.Fatalf("pricing formula should render as a compact table header tooltip")
	}
	for _, want := range []string{"Cache Read", "Cache Write", "Reasoning Tokens", "Tool Tokens"} {
		if !strings.Contains(indexResp.Body.String(), want) {
			t.Fatalf("index body does not contain extended token fragment %q", want)
		}
	}

	apiReq := httptest.NewRequest(http.MethodGet, "/api/dashboard?grain=day", nil)
	apiResp := httptest.NewRecorder()
	server.Routes().ServeHTTP(apiResp, apiReq)
	if apiResp.Code != http.StatusOK {
		t.Fatalf("api status = %d, want 200", apiResp.Code)
	}
	if !strings.Contains(apiResp.Body.String(), `"total_tokens":120`) {
		t.Fatalf("api body = %s, want total_tokens 120", apiResp.Body.String())
	}
}

func TestEmbeddedServerRendersIndexAndStaticAssets(t *testing.T) {
	server, err := NewEmbeddedServer([]usage.Event{
		{
			ToolName:    "codex",
			ModelName:   "gpt-5",
			InputTokens: 100,
			TotalTokens: 100,
			OccurredAt:  time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	indexResp := httptest.NewRecorder()
	server.Routes().ServeHTTP(indexResp, httptest.NewRequest(http.MethodGet, "/", nil))
	if indexResp.Code != http.StatusOK {
		t.Fatalf("index status = %d, want 200", indexResp.Code)
	}
	if !strings.Contains(indexResp.Body.String(), "AgentMeter") {
		t.Fatalf("index body does not contain AgentMeter")
	}

	staticResp := httptest.NewRecorder()
	server.Routes().ServeHTTP(staticResp, httptest.NewRequest(http.MethodGet, "/static/css/app.css", nil))
	if staticResp.Code != http.StatusOK {
		t.Fatalf("static status = %d, want 200", staticResp.Code)
	}
	if !strings.Contains(staticResp.Body.String(), "--bg:") {
		t.Fatalf("embedded static asset was not served")
	}
}

func TestIndexModelTableAlwaysShowsDailyModelRows(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{
			ToolName:     "codex",
			ModelName:    "gpt-5",
			InputTokens:  100,
			OutputTokens: 20,
			TotalTokens:  120,
			OccurredAt:   time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ToolName:     "codex",
			ModelName:    "gpt-5",
			InputTokens:  200,
			OutputTokens: 30,
			TotalTokens:  230,
			OccurredAt:   time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC),
		},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	monthResp := httptest.NewRecorder()
	server.Routes().ServeHTTP(monthResp, httptest.NewRequest(http.MethodGet, "/?grain=month", nil))
	if !strings.Contains(monthResp.Body.String(), ">2026-05-01<") || !strings.Contains(monthResp.Body.String(), ">2026-05-02<") {
		t.Fatalf("model table should keep daily periods even when trend grain is month:\n%s", monthResp.Body.String())
	}
	if strings.Index(monthResp.Body.String(), ">2026-05-02<") > strings.Index(monthResp.Body.String(), ">2026-05-01<") {
		t.Fatalf("model table should show latest day first:\n%s", monthResp.Body.String())
	}
}

func TestIndexModelTableUsesCodexBillableEstimate(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{
			ToolName:        "codex",
			ModelName:       "gpt-5.5",
			InputTokens:     128_415_423,
			OutputTokens:    302_708,
			CacheReadTokens: 124_555_776,
			TotalTokens:     128_718_131,
			OccurredAt:      time.Date(2026, 5, 27, 14, 4, 30, 0, time.UTC),
		},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/", nil))
	body := resp.Body.String()

	if !strings.Contains(body, ">2026-05-27<") || !strings.Contains(body, "$65.116") {
		t.Fatalf("codex model row should use billable estimate near $65:\n%s", body)
	}
}

func TestModelTableSeparatesSameModelByTool(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{
			ToolName:    "codex",
			ModelName:   "gpt-5",
			InputTokens: 100,
			TotalTokens: 100,
			OccurredAt:  time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			ToolName:    "cursor",
			ModelName:   "gpt-5",
			InputTokens: 200,
			TotalTokens: 200,
			OccurredAt:  time.Date(2026, 5, 2, 11, 0, 0, 0, time.UTC),
		},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/", nil))
	body := resp.Body.String()

	if !strings.Contains(body, "<th>Tool</th>") {
		t.Fatalf("model table should include Tool column:\n%s", body)
	}
	if strings.Count(body, "<td>gpt-5</td>") != 2 {
		t.Fatalf("same model used by different tools should render separate rows:\n%s", body)
	}
	if !strings.Contains(body, "<td>Codex</td>") || !strings.Contains(body, "<td>Cursor</td>") {
		t.Fatalf("model table should render tool names:\n%s", body)
	}
}

func TestModelTableDoesNotRenderUnknownWhenScannerResolvedModelContext(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{
			ToolName:    "codex",
			ModelName:   "gpt-5-codex",
			InputTokens: 100,
			TotalTokens: 100,
			OccurredAt:  time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
		},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/", nil))
	body := resp.Body.String()

	if !strings.Contains(body, "<td>gpt-5-codex</td>") {
		t.Fatalf("model table should render the resolved model:\n%s", body)
	}
	if strings.Contains(body, "<td>unknown</td>") {
		t.Fatalf("model table should not render unknown when model was resolved:\n%s", body)
	}
}

func TestIndexSupportsEnglishLanguage(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 100, TotalTokens: 100, OccurredAt: time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/?lang=en", nil))
	body := resp.Body.String()

	if !strings.Contains(body, "Local Agent Usage") || !strings.Contains(body, "Daily API Requests") || !strings.Contains(body, "Export") {
		t.Fatalf("english page missing translated labels:\n%s", body)
	}
	if strings.Contains(body, "本地 Agent 用量统计") || strings.Contains(body, "每日 API 请求次数") {
		t.Fatalf("english page should not show Chinese labels:\n%s", body)
	}
}

func TestIndexDefaultsToEnglishAndUsesClickLanguageToggle(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 100, TotalTokens: 100, OccurredAt: time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/", nil))
	body := resp.Body.String()

	if !strings.Contains(body, `<html lang="en">`) || !strings.Contains(body, "Local Agent Usage") {
		t.Fatalf("default page should render in English:\n%s", body)
	}
	if !strings.Contains(body, `class="page-controls"`) || !strings.Contains(body, `class="language-switch`) || !strings.Contains(body, `data-next-lang="zh-CN"`) {
		t.Fatalf("default page should render a top-right language switch to Chinese:\n%s", body)
	}
	if !strings.Contains(body, `<span>EN</span>`) || !strings.Contains(body, `<strong>中文</strong>`) {
		t.Fatalf("language switch should render like a two-state switch:\n%s", body)
	}
	if strings.Contains(body, `id="langSelect"`) {
		t.Fatalf("language switcher should not be a select dropdown:\n%s", body)
	}
}

func TestWebDashboardSupportsThemeToggle(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 100, TotalTokens: 100, OccurredAt: time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/", nil))
	body := resp.Body.String()

	for _, want := range []string{`id="themeToggle"`, `data-theme-toggle`, `aria-label="Theme"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("index should render a theme toggle fragment %q:\n%s", want, body)
		}
	}

	style, err := os.ReadFile(filepath.Join(root, "static", "css", "app.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(style), `body[data-theme="light"]`) {
		t.Fatalf("stylesheet should define light theme variables")
	}
	if !strings.Contains(string(style), ".page-controls") || strings.Contains(string(style), ".floating-controls") {
		t.Fatalf("global controls should live in the page header instead of a floating rail")
	}

	script, err := os.ReadFile(filepath.Join(root, "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"agentmeter-theme", "setupThemeToggle", "document.body.dataset.theme"} {
		if !strings.Contains(string(script), want) {
			t.Fatalf("theme toggle script missing %q", want)
		}
	}
}

func TestIndexSupportsChineseLanguage(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 100, TotalTokens: 100, OccurredAt: time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/?lang=zh-CN", nil))
	body := resp.Body.String()

	if !strings.Contains(body, "本地 Agent 用量统计") || !strings.Contains(body, "每日 API 请求次数") || !strings.Contains(body, "<th>工具</th>") || !strings.Contains(body, "费用估算") {
		t.Fatalf("chinese page missing translated labels:\n%s", body)
	}
	if !strings.Contains(body, `class="language-switch`) || !strings.Contains(body, `data-next-lang="en"`) {
		t.Fatalf("chinese page should render a top-right language switch to English:\n%s", body)
	}
	if !strings.Contains(body, `<strong>EN</strong>`) || !strings.Contains(body, `<span>中文</span>`) {
		t.Fatalf("language switch should render like a two-state switch:\n%s", body)
	}
}

func TestTrendChartLabelsDoNotShowTokenValueBelowDate(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("..", "..", "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(script)

	if strings.Contains(body, "label.innerHTML") && strings.Contains(body, "<strong>${formatCompact(item.total_tokens)}</strong>") {
		t.Fatalf("trend chart date labels should not render token values below the date")
	}
	if !strings.Contains(body, "formatNumber(item.total_tokens)} tokens") {
		t.Fatalf("trend chart should keep the full token value in the hover title")
	}
}

func TestFrontendNumberFormattingFollowsPageLanguage(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("..", "..", "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(script)

	if strings.Contains(body, `Intl.NumberFormat("zh-CN"`) {
		t.Fatalf("frontend number formatting should not hardcode zh-CN")
	}
	if !strings.Contains(body, "document.documentElement.lang") || !strings.Contains(body, "zh-CN") || !strings.Contains(body, "en") {
		t.Fatalf("frontend number formatting should derive locale from the html lang")
	}
}

func TestProjectRankingDisplaysProjectNameOnly(t *testing.T) {
	template, err := os.ReadFile(filepath.Join("..", "..", "templates", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(filepath.Join("..", "..", "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(template), `data-ranking-kind="project"`) {
		t.Fatalf("project ranking should be marked so the frontend can format project names")
	}
	if !strings.Contains(string(script), "formatRankingName") || !strings.Contains(string(script), `kind === "project"`) {
		t.Fatalf("frontend should format project ranking labels as project names")
	}
}

func TestProjectRankingIsScrollableAndNotLimitedToTen(t *testing.T) {
	template, err := os.ReadFile(filepath.Join("..", "..", "templates", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(filepath.Join("..", "..", "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	style, err := os.ReadFile(filepath.Join("..", "..", "static", "css", "app.css"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(template), `class="ranking project-ranking"`) {
		t.Fatalf("project ranking should have a dedicated scrollable class")
	}
	if !strings.Contains(string(script), `kind === "project" ? data : data.slice(0, 10)`) {
		t.Fatalf("project ranking should render all rows while other rankings stay limited")
	}
	if !strings.Contains(string(style), ".project-ranking") || !strings.Contains(string(style), "overflow-y: auto") {
		t.Fatalf("project ranking should be vertically scrollable")
	}
}

func TestDailyRequestChartShowsRequestCountsAtEachPoint(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("..", "..", "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	style, err := os.ReadFile(filepath.Join("..", "..", "static", "css", "app.css"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(script), "line-point-value") || !strings.Contains(string(script), "formatNumber(item.request_count)") {
		t.Fatalf("daily request chart should render request count labels at each point")
	}
	if !strings.Contains(string(style), ".line-point-value") {
		t.Fatalf("daily request chart point labels should have CSS styling")
	}
}

func TestModelUsageDetailsSupportsToolAndModelFilters(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 100, OutputTokens: 20, TotalTokens: 120, OccurredAt: time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)},
		{ToolName: "cursor", ModelName: "claude-sonnet-4", InputTokens: 80, OutputTokens: 10, TotalTokens: 90, OccurredAt: time.Date(2026, 5, 2, 11, 0, 0, 0, time.UTC)},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/", nil))
	body := resp.Body.String()

	for _, want := range []string{`id="modelToolFilter"`, `id="modelNameFilter"`, `data-tool="Codex"`, `data-model="gpt-5"`, `data-total-tokens="120"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("model details filter markup missing %q:\n%s", want, body)
		}
	}

	script, err := os.ReadFile(filepath.Join(root, "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"setupModelFilters", "applyModelFilters", "recalculateModelTotal", "modelToolFilter", "modelNameFilter"} {
		if !strings.Contains(string(script), want) {
			t.Fatalf("model details filter script missing %q", want)
		}
	}
}

func TestWebDashboardSupportsGlobalTokenDisplayToggle(t *testing.T) {
	root := filepath.Join("..", "..")
	server, err := NewServer([]usage.Event{
		{ToolName: "codex", ModelName: "gpt-5", InputTokens: 7454012675, OutputTokens: 25990787, TotalTokens: 7704105873, OccurredAt: time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)},
	}, filepath.Join(root, "templates"), filepath.Join(root, "static"))
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	server.Routes().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/", nil))
	body := resp.Body.String()

	for _, want := range []string{`id="tokenDisplayToggle"`, `data-token-mode-toggle`, `data-token-mode="compact"`, `data-token-value="7454012675"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("global token display toggle markup missing %q:\n%s", want, body)
		}
	}

	script, err := os.ReadFile(filepath.Join(root, "static", "js", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"agentmeter-token-display", "setupTokenDisplayToggle", "applyTokenDisplayMode", "formatCompact"} {
		if !strings.Contains(string(script), want) {
			t.Fatalf("token display toggle script missing %q", want)
		}
	}
}
