package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunShowsHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr.String())
	}
	for _, want := range []string{"Usage:", "summary", "serve", "scan", "doctor", "--version"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("help output missing %q:\n%s", want, stdout.String())
		}
	}
}

func TestRunShowsVersion(t *testing.T) {
	oldVersion := version
	version = "test-version"
	defer func() { version = oldVersion }()

	var stdout, stderr bytes.Buffer
	code := run([]string{"-v"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "agentmeter test-version" {
		t.Fatalf("version output = %q", stdout.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"missing"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown command: missing") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
