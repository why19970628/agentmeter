package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/why19970628/agentmeter/internal/usage"
)

type Options struct {
	Paths string
}

func Report(opts Options) string {
	var b strings.Builder
	fmt.Fprintln(&b, "AgentMeter Doctor")
	for _, item := range strings.Split(opts.Paths, ",") {
		raw := strings.TrimSpace(item)
		if raw == "" {
			continue
		}
		path := usage.ExpandHome(raw)
		fmt.Fprintf(&b, "\npath: %s\n", displayPath(path, raw))

		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintln(&b, "status: missing")
			continue
		}
		if !info.IsDir() {
			fmt.Fprintln(&b, "status: not a directory")
			continue
		}

		files := countUsageFiles(path)
		events, err := usage.ScanDir(path, usage.ScanOptions{})
		if err != nil {
			fmt.Fprintln(&b, "status: scan error")
			continue
		}
		dashboard := usage.BuildDashboard(events, usage.GrainDay)
		fmt.Fprintln(&b, "status: ok")
		fmt.Fprintf(&b, "usage files: %d\n", files)
		fmt.Fprintf(&b, "events: %d\n", len(events))
		fmt.Fprintf(&b, "tools: %d\n", dashboard.Summary.ToolCount)
		fmt.Fprintf(&b, "models: %d\n", dashboard.Summary.ModelCount)
		fmt.Fprintf(&b, "tokens: %d\n", dashboard.Summary.TotalTokens)
		fmt.Fprintf(&b, "latest event: %s\n", latestEventTime(events))
		if len(events) > 0 {
			fmt.Fprintf(&b, "detected tools: %s\n", strings.Join(uniqueTools(events), ", "))
		}
	}
	return b.String()
}

func countUsageFiles(root string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if usage.LooksLikeUsageFile(path) {
			count++
		}
		return nil
	})
	return count
}

func latestEventTime(events []usage.Event) string {
	var latest time.Time
	for _, event := range events {
		if event.OccurredAt.After(latest) {
			latest = event.OccurredAt
		}
	}
	if latest.IsZero() {
		return "none"
	}
	return latest.Format(time.RFC3339)
}

func uniqueTools(events []usage.Event) []string {
	set := map[string]struct{}{}
	for _, event := range events {
		name := strings.TrimSpace(event.ToolName)
		if name == "" {
			name = "unknown"
		}
		set[name] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func displayPath(expanded string, raw string) string {
	if strings.HasPrefix(raw, "~/") || raw == "~" {
		return raw
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		if expanded == home {
			return "~"
		}
		if strings.HasPrefix(expanded, home+string(os.PathSeparator)) {
			return "~/" + strings.TrimPrefix(expanded, home+string(os.PathSeparator))
		}
	}
	return raw
}
