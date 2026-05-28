package usage

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultCodexDir  = ".codex"
	DefaultClaudeDir = ".claude"
	DefaultCursorDir = ".cursor"
	DefaultGeminiDir = ".gemini"

	UsageExtJSON  = ".json"
	UsageExtJSONL = ".jsonl"
	UsageExtLog   = ".log"
)

var DefaultDirNames = []string{
	DefaultCodexDir,
	DefaultClaudeDir,
	DefaultCursorDir,
	DefaultGeminiDir,
}

func DefaultPaths() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	paths := make([]string, 0, len(DefaultDirNames))
	for _, name := range DefaultDirNames {
		paths = append(paths, filepath.Join(home, name))
	}
	return strings.Join(paths, ",")
}

func ScanPaths(raw string, opts ScanOptions) []Event {
	var events []Event
	for _, path := range SplitPaths(raw) {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		found, err := ScanDir(path, opts)
		if err != nil {
			continue
		}
		events = append(events, found...)
	}
	return NormalizeCumulativeEvents(events)
}

func SplitPaths(raw string) []string {
	items := strings.Split(raw, ",")
	paths := make([]string, 0, len(items))
	for _, item := range items {
		path := ExpandHome(strings.TrimSpace(item))
		if path == "" {
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

func ExpandHome(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func LooksLikeUsageFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case UsageExtJSON, UsageExtJSONL, UsageExtLog:
		return true
	default:
		return false
	}
}
