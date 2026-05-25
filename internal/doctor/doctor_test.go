package doctor

import (
	"strings"
	"testing"
)

func TestReportShowsSourceHealth(t *testing.T) {
	report := Report(Options{Paths: "../../examples"})

	for _, want := range []string{
		"AgentMeter Doctor",
		"path: ../../examples",
		"status: ok",
		"usage files:",
		"events:",
		"tools:",
		"models:",
		"latest event:",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("doctor report missing %q:\n%s", want, report)
		}
	}
}
