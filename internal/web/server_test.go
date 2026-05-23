package web

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agentmeter/internal/usage"
)

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
	if !strings.Contains(indexResp.Body.String(), "每日 API 请求次数") {
		t.Fatalf("index body does not contain daily request chart title")
	}
	if !strings.Contains(indexResp.Body.String(), "request_count") {
		t.Fatalf("index body does not expose daily request count")
	}
	for _, want := range []string{"模型用量明细", "Model", "Input Tokens", "Output Tokens", "Total Tokens", "Est. USD", "gpt-5"} {
		if !strings.Contains(indexResp.Body.String(), want) {
			t.Fatalf("index body does not contain model usage table fragment %q", want)
		}
	}
	for _, want := range []string{"缓存读取", "缓存写入", "推理 Tokens", "Cache Read", "Cache Write", "Reasoning", "Tool Tokens"} {
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
