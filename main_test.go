package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveWebAssetsFindsDirectoryWithTemplatesAndStatic(t *testing.T) {
	root := t.TempDir()
	assets := filepath.Join(root, "agentmeter")
	if err := os.MkdirAll(filepath.Join(assets, "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(assets, "static"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "templates", "layout.html"), []byte("layout"), 0o644); err != nil {
		t.Fatal(err)
	}

	templateDir, staticDir, ok := resolveWebAssets([]string{filepath.Join(root, "missing"), assets})
	if !ok {
		t.Fatal("resolveWebAssets did not find asset directory")
	}
	if templateDir != filepath.Join(assets, "templates") {
		t.Fatalf("templateDir = %q, want %q", templateDir, filepath.Join(assets, "templates"))
	}
	if staticDir != filepath.Join(assets, "static") {
		t.Fatalf("staticDir = %q, want %q", staticDir, filepath.Join(assets, "static"))
	}
}
