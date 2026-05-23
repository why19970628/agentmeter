package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agentmeter/internal/report"
	"agentmeter/internal/usage"
	"agentmeter/internal/web"
)

func main() {
	if len(os.Args) < 2 {
		serve(os.Args[1:])
		return
	}

	switch os.Args[1] {
	case "summary":
		summary(os.Args[2:])
	case "serve":
		serve(os.Args[2:])
	case "scan":
		scan(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(2)
	}
}

func summary(args []string) {
	fs := flag.NewFlagSet("summary", flag.ExitOnError)
	paths := fs.String("paths", defaultPaths(), "comma-separated local log directories")
	period := fs.String("period", "month", "period: today, week, month, all")
	format := fs.String("format", "table", "format: table, json, markdown")
	group := fs.String("group", "source", "group: source, model")
	lang := fs.String("lang", "en", "language: en, zh-CN")
	_ = fs.Parse(args)

	events := scanPaths(*paths)
	events = report.FilterPeriod(events, *period, time.Now())
	fmt.Print(report.Render(events, report.Options{Format: *format, Group: *group, Lang: *lang}))
}

func serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8787", "listen address")
	paths := fs.String("paths", defaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	events := scanPaths(*paths)
	server, err := web.NewServer(events, "templates", "static")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("AgentMeter loaded %d events", len(events))
	log.Printf("AgentMeter is running at http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Routes()))
}

func scan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	paths := fs.String("paths", defaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	events := scanPaths(*paths)
	dashboard := usage.BuildDashboard(events, usage.GrainDay)
	fmt.Printf("events: %d\n", len(events))
	fmt.Printf("tools: %d\n", dashboard.Summary.ToolCount)
	fmt.Printf("models: %d\n", dashboard.Summary.ModelCount)
	fmt.Printf("tokens: %d\n", dashboard.Summary.TotalTokens)
}

func scanPaths(raw string) []usage.Event {
	var events []usage.Event
	for _, item := range strings.Split(raw, ",") {
		path := expandHome(strings.TrimSpace(item))
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}
		found, err := usage.ScanDir(path, usage.ScanOptions{})
		if err != nil {
			log.Printf("scan %s failed: %v", path, err)
			continue
		}
		events = append(events, found...)
	}
	return events
}

func defaultPaths() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return strings.Join([]string{
		filepath.Join(home, ".codex"),
		filepath.Join(home, ".claude"),
		filepath.Join(home, ".cursor"),
		filepath.Join(home, ".gemini"),
	}, ",")
}

func expandHome(path string) string {
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
