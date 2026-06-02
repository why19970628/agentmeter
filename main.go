package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/why19970628/agentmeter/internal/doctor"
	"github.com/why19970628/agentmeter/internal/report"
	"github.com/why19970628/agentmeter/internal/usage"
	"github.com/why19970628/agentmeter/internal/web"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		serve([]string{}, stdout, stderr)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printHelp(stdout)
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "agentmeter %s\n", version)
	case "summary":
		summary(args[1:], stdout, stderr)
	case "serve":
		serve(args[1:], stdout, stderr)
	case "scan":
		scan(args[1:], stdout, stderr)
	case "doctor":
		runDoctor(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return 2
	}
	return 0
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, `AgentMeter - offline-first local AI agent usage tracker

Usage:
  agentmeter [command] [flags]

Commands:
  serve     Start the web dashboard (default)
  summary   Print usage summary
  scan      Scan local usage logs and print counters
  doctor    Check configured usage paths

Flags:
  -h, --help      Show help
  -v, --version   Show version

Examples:
  agentmeter serve
  agentmeter summary -period all -group tool -group model
  agentmeter scan -paths ~/.codex,~/.claude
`)
}

func summary(args []string, stdout io.Writer, stderr io.Writer) {
	fs := flag.NewFlagSet("summary", flag.ExitOnError)
	fs.SetOutput(stderr)
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	period := fs.String("period", "month", "period: today, week, month, all")
	format := fs.String("format", "table", "format: table, json, markdown")
	var groups multiFlag
	fs.Var(&groups, "group", "group: source, tool, model, tool,model. Can be repeated")
	lang := fs.String("lang", "en", "language: en, zh-CN")
	_ = fs.Parse(args)

	events := usage.ScanPaths(*paths, usage.ScanOptions{})
	events = report.FilterPeriod(events, *period, time.Now())
	fmt.Fprint(stdout, report.Render(events, report.Options{Format: *format, Group: report.NormalizeGroup(groups), Lang: *lang}))
}

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func serve(args []string, stdout io.Writer, stderr io.Writer) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.SetOutput(stderr)
	addr := fs.String("addr", "127.0.0.1:8787", "listen address")
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	events := usage.ScanPaths(*paths, usage.ScanOptions{})
	templateDir, staticDir, ok := resolveWebAssets(defaultWebAssetCandidates())
	if !ok {
		log.Print("AgentMeter web assets not found on disk; using embedded assets")
	}
	server, err := web.NewServer(events, templateDir, staticDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("AgentMeter loaded %d events", len(events))
	log.Printf("AgentMeter is running at http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Routes()))
}

func scan(args []string, stdout io.Writer, stderr io.Writer) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	fs.SetOutput(stderr)
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	events := usage.ScanPaths(*paths, usage.ScanOptions{})
	dashboard := usage.BuildDashboard(events, usage.GrainDay)
	fmt.Fprintf(stdout, "events: %d\n", len(events))
	fmt.Fprintf(stdout, "tools: %d\n", dashboard.Summary.ToolCount)
	fmt.Fprintf(stdout, "models: %d\n", dashboard.Summary.ModelCount)
	fmt.Fprintf(stdout, "tokens: %d\n", dashboard.Summary.TotalTokens)
}

func runDoctor(args []string, stdout io.Writer, stderr io.Writer) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	fs.SetOutput(stderr)
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	fmt.Fprint(stdout, doctor.Report(doctor.Options{Paths: *paths}))
}

func defaultWebAssetCandidates() []string {
	candidates := []string{}
	if env := strings.TrimSpace(os.Getenv("AGENTMETER_ASSET_DIR")); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, ".")
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			exeDir,
			filepath.Join(exeDir, ".."),
			filepath.Join(exeDir, "..", "share", "agentmeter"),
			filepath.Join(exeDir, "..", "share", "agentmeter", "libexec", "agentmeter"),
		)
	}
	candidates = append(candidates,
		"/usr/local/share/agentmeter",
		"/opt/homebrew/share/agentmeter",
	)
	return candidates
}

func resolveWebAssets(candidates []string) (string, string, bool) {
	for _, candidate := range candidates {
		root := usage.ExpandHome(strings.TrimSpace(candidate))
		if root == "" {
			continue
		}
		templateDir := filepath.Join(root, "templates")
		staticDir := filepath.Join(root, "static")
		if fileExists(filepath.Join(templateDir, "layout.html")) && dirExists(staticDir) {
			return templateDir, staticDir, true
		}
	}
	return "", "", false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
