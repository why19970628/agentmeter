package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirParsesJSONLUsageRecords(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	body := `{"tool":"codex","model":"gpt-5","session_id":"s1","project":"/work/a","usage":{"input_tokens":120,"output_tokens":30},"timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool_name":"claude-code","model_name":"claude-sonnet","input_tokens":50,"output_tokens":20,"occurred_at":"2026-05-04T11:00:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{DefaultTool: "generic"})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0].ToolName != "codex" || events[0].ModelName != "gpt-5" {
		t.Fatalf("first event identity = %+v, want codex gpt-5", events[0])
	}
	if events[0].TotalTokens != 150 {
		t.Fatalf("first event total = %d, want 150", events[0].TotalTokens)
	}
	if events[1].ToolName != "claude-code" || events[1].TotalTokens != 70 {
		t.Fatalf("second event = %+v, want claude-code total 70", events[1])
	}
}

func TestScanDirSkipsEmptyJSONFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "empty.json"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("events len = %d, want 0", len(events))
	}
}

func TestScanDirParsesNestedModelFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested.jsonl")
	body := `{"tool":"codex","message":{"model":"gpt-5","usage":{"input_tokens":90,"output_tokens":10}},"timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool":"claude-code","request":{"model":"claude-sonnet-4"},"usage":{"input_tokens":30,"output_tokens":20},"timestamp":"2026-05-03T11:00:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0].ModelName != "gpt-5" {
		t.Fatalf("first model = %q, want gpt-5", events[0].ModelName)
	}
	if events[0].TotalTokens != 100 {
		t.Fatalf("first total tokens = %d, want 100", events[0].TotalTokens)
	}
	if events[1].ModelName != "claude-sonnet-4" {
		t.Fatalf("second model = %q, want claude-sonnet-4", events[1].ModelName)
	}
}

func TestScanDirParsesCommonModelAliases(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model-aliases.jsonl")
	body := `{"tool":"codex","message":{"model_slug":"gpt-5-codex"},"usage":{"input_tokens":90,"output_tokens":10},"timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool":"claude-code","request":{"modelId":"claude-sonnet-4"},"usage":{"input_tokens":30,"output_tokens":20},"timestamp":"2026-05-03T11:00:00Z"}` + "\n" +
		`{"tool":"gemini","response":{"model_id":"gemini-2.5-pro"},"usage":{"prompt_tokens":20,"completion_tokens":10},"timestamp":"2026-05-03T12:00:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 3 {
		t.Fatalf("events len = %d, want 3", len(events))
	}
	for idx, want := range []string{"gpt-5-codex", "claude-sonnet-4", "gemini-2.5-pro"} {
		if events[idx].ModelName != want {
			t.Fatalf("event %d model = %q, want %q", idx, events[idx].ModelName, want)
		}
	}
}

func TestScanDirCarriesModelContextWithinJSONLFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context.jsonl")
	body := `{"tool":"codex","session_id":"s1","model":"gpt-5","timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool":"codex","session_id":"s1","usage":{"input_tokens":12,"output_tokens":8},"timestamp":"2026-05-03T10:01:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].ModelName != "gpt-5" {
		t.Fatalf("event model = %q, want gpt-5", events[0].ModelName)
	}
}

func TestScanDirCarriesLaterModelContextWithinJSONLFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "later-context.jsonl")
	body := `{"tool":"codex","session_id":"s1","usage":{"input_tokens":496,"output_tokens":258,"output_tokens_details":{"reasoning_tokens":40}},"timestamp":"2026-05-24T10:01:00Z"}` + "\n" +
		`{"tool":"codex","session_id":"s1","model":"gpt-5.5","timestamp":"2026-05-24T10:02:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].ModelName != "gpt-5.5" {
		t.Fatalf("event model = %q, want gpt-5.5", events[0].ModelName)
	}
}

func TestScanDirCarriesModelContextWithinJSONDocument(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context.json")
	body := `[
		{"tool":"codex","session_id":"s1","model":"gpt-5","timestamp":"2026-05-03T10:00:00Z"},
		{"tool":"codex","session_id":"s1","usage":{"input_tokens":12,"output_tokens":8},"timestamp":"2026-05-03T10:01:00Z"}
	]`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].ModelName != "gpt-5" {
		t.Fatalf("event model = %q, want gpt-5", events[0].ModelName)
	}
}

func TestScanDirCarriesLaterModelContextWithinJSONDocument(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "later-context.json")
	body := `[
		{"tool":"codex","session_id":"s1","usage":{"input_tokens":496,"output_tokens":258,"output_tokens_details":{"reasoning_tokens":40}},"timestamp":"2026-05-24T10:01:00Z"},
		{"tool":"codex","session_id":"s1","model":"gpt-5.5","timestamp":"2026-05-24T10:02:00Z"}
	]`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].ModelName != "gpt-5.5" {
		t.Fatalf("event model = %q, want gpt-5.5", events[0].ModelName)
	}
}

func TestScanDirParsesCommonProjectAliases(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "project-aliases.jsonl")
	body := `{"tool":"codex","model":"gpt-5","workspace_path":"/work/agentmeter","usage":{"input_tokens":90,"output_tokens":10},"timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool":"claude-code","model":"claude-sonnet-4","message":{"current_working_directory":"/work/qcds"},"usage":{"input_tokens":30,"output_tokens":20},"timestamp":"2026-05-03T11:00:00Z"}` + "\n" +
		`{"tool":"gemini","model":"gemini-2.5-pro","request":{"root_path":"/work/gemini"},"usage":{"prompt_tokens":20,"completion_tokens":10},"timestamp":"2026-05-03T12:00:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 3 {
		t.Fatalf("events len = %d, want 3", len(events))
	}
	for idx, want := range []string{"/work/agentmeter", "/work/qcds", "/work/gemini"} {
		if events[idx].ProjectPath != want {
			t.Fatalf("event %d project = %q, want %q", idx, events[idx].ProjectPath, want)
		}
	}
}

func TestScanDirCarriesProjectContextWithinJSONLFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "project-context.jsonl")
	body := `{"tool":"codex","session_id":"s1","model":"gpt-5","cwd":"/work/agentmeter","timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool":"codex","session_id":"s1","usage":{"input_tokens":12,"output_tokens":8},"timestamp":"2026-05-03T10:01:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].ProjectPath != "/work/agentmeter" {
		t.Fatalf("event project = %q, want /work/agentmeter", events[0].ProjectPath)
	}
}

func TestScanDirParsesCacheReasoningAndToolTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "extended.jsonl")
	body := `{"tool":"claude-code","model":"claude-sonnet-4","usage":{"input_tokens":100,"output_tokens":40,"cache_read_input_tokens":20,"cache_creation_input_tokens":10,"reasoning_tokens":7,"tool_tokens":3},"timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool":"gemini","model":"gemini-2.5-pro","usage":{"prompt_tokens":60,"completion_tokens":30,"cached_content_token_count":11,"thoughts_token_count":9},"timestamp":"2026-05-03T11:00:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0].CacheReadTokens != 20 || events[0].CacheWriteTokens != 10 || events[0].ReasoningTokens != 7 || events[0].ToolTokens != 3 {
		t.Fatalf("first event extended tokens = %+v", events[0])
	}
	if events[0].TotalTokens != 160 {
		t.Fatalf("first event total = %d, want 160", events[0].TotalTokens)
	}
	if events[1].CacheReadTokens != 11 || events[1].ReasoningTokens != 9 || events[1].TotalTokens != 99 {
		t.Fatalf("second event extended tokens = %+v, want cache read 11 reasoning 9 total 99", events[1])
	}
}

func TestScanDirParsesVendorSpecificTokenDetails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vendor-details.jsonl")
	body := `{"tool":"claude-code","model":"claude-sonnet-4","usage":{"input_tokens":100,"output_tokens":40,"cache_read_input_tokens":20,"cache_creation":{"ephemeral_5m_input_tokens":6,"ephemeral_1h_input_tokens":4}},"timestamp":"2026-05-03T10:00:00Z"}` + "\n" +
		`{"tool":"codex","model":"gpt-5","usage":{"input_tokens":80,"output_tokens":30,"input_tokens_details":{"cached_tokens":15},"output_tokens_details":{"reasoning_tokens":12},"tool_tokens":5},"timestamp":"2026-05-03T11:00:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := ScanDir(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0].CacheWriteTokens != 10 {
		t.Fatalf("claude cache write = %d, want 10", events[0].CacheWriteTokens)
	}
	if events[1].CacheReadTokens != 15 || events[1].ReasoningTokens != 12 || events[1].ToolTokens != 5 {
		t.Fatalf("codex details = %+v, want cache read 15 reasoning 12 tool 5", events[1])
	}
}
