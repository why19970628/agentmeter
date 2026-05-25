package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"~":               home,
		"~/agentmeter":    filepath.Join(home, "agentmeter"),
		"/tmp/agentmeter": "/tmp/agentmeter",
	}

	for input, want := range tests {
		if got := ExpandHome(input); got != want {
			t.Fatalf("ExpandHome(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestScanPathsReadsCommaSeparatedDirectories(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	missing := filepath.Join(t.TempDir(), "missing")

	writeUsageFixture(t, first, "one.jsonl", `{"timestamp":"2026-05-01T00:00:00Z","model":"gpt-5.5","usage":{"input_tokens":10,"output_tokens":5}}`)
	writeUsageFixture(t, second, "two.jsonl", `{"timestamp":"2026-05-02T00:00:00Z","model":"claude-sonnet","usage":{"input_tokens":7,"output_tokens":3}}`)

	events := ScanPaths(first+","+missing+","+second, ScanOptions{})
	if len(events) != 2 {
		t.Fatalf("ScanPaths returned %d events, want 2", len(events))
	}
}

func writeUsageFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
