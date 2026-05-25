package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/why19970628/agentmeter/internal/doctor"
	"github.com/why19970628/agentmeter/internal/report"
	"github.com/why19970628/agentmeter/internal/usage"
	"github.com/why19970628/agentmeter/internal/web"
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
	case "doctor":
		runDoctor(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(2)
	}
}

func summary(args []string) {
	fs := flag.NewFlagSet("summary", flag.ExitOnError)
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	period := fs.String("period", "month", "period: today, week, month, all")
	format := fs.String("format", "table", "format: table, json, markdown")
	var groups multiFlag
	fs.Var(&groups, "group", "group: source, tool, model, tool,model. Can be repeated")
	lang := fs.String("lang", "en", "language: en, zh-CN")
	_ = fs.Parse(args)

	events := usage.ScanPaths(*paths, usage.ScanOptions{})
	events = report.FilterPeriod(events, *period, time.Now())
	fmt.Print(report.Render(events, report.Options{Format: *format, Group: report.NormalizeGroup(groups), Lang: *lang}))
}

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8787", "listen address")
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	events := usage.ScanPaths(*paths, usage.ScanOptions{})
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
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	events := usage.ScanPaths(*paths, usage.ScanOptions{})
	dashboard := usage.BuildDashboard(events, usage.GrainDay)
	fmt.Printf("events: %d\n", len(events))
	fmt.Printf("tools: %d\n", dashboard.Summary.ToolCount)
	fmt.Printf("models: %d\n", dashboard.Summary.ModelCount)
	fmt.Printf("tokens: %d\n", dashboard.Summary.TotalTokens)
}

func runDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	paths := fs.String("paths", usage.DefaultPaths(), "comma-separated local log directories")
	_ = fs.Parse(args)

	fmt.Print(doctor.Report(doctor.Options{Paths: *paths}))
}
