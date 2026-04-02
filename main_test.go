package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/karldreher/gh-qpr/lib"
)

func writeConfig(t *testing.T, home string, cfg lib.Config) {
	t.Helper()
	dir := filepath.Join(home, ".gh-qpr")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveTemplate_FlagTakesPriority(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("GH_QPR_DEFAULT_TEMPLATE", "from-env")
	writeConfig(t, tmpDir, lib.Config{DefaultTemplate: "from-config"})

	got, err := resolveTemplate("explicit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "explicit" {
		t.Errorf("expected %q, got %q", "explicit", got)
	}
}

func TestResolveTemplate_EnvVarOverridesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("GH_QPR_DEFAULT_TEMPLATE", "from-env")
	writeConfig(t, tmpDir, lib.Config{DefaultTemplate: "from-config"})

	got, err := resolveTemplate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from-env" {
		t.Errorf("expected %q, got %q", "from-env", got)
	}
}

func TestResolveTemplate_ConfigFallback(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("GH_QPR_DEFAULT_TEMPLATE", "")
	writeConfig(t, tmpDir, lib.Config{DefaultTemplate: "from-config"})

	got, err := resolveTemplate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from-config" {
		t.Errorf("expected %q, got %q", "from-config", got)
	}
}

// setupFakeRepo creates a fake template repo cache under HOME and returns the
// templates directory path. It sets HOME and GH_QPR_REPO via t.Setenv.
func setupFakeRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("GH_QPR_REPO", "testowner/testrepo")
	templatesDir := filepath.Join(tmpDir, ".gh-qpr", "testrepo", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}
	return templatesDir
}

// captureStdout redirects os.Stdout, calls f, then returns what was written.
// os.Stdout is restored via defer so teardown is guaranteed even on panic.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()
	f()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRunListCmd_PrintsTemplateNames(t *testing.T) {
	templatesDir := setupFakeRepo(t)
	for _, name := range []string{"alpha.md", "beta.md"} {
		if err := os.WriteFile(filepath.Join(templatesDir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	output := captureStdout(t, func() {
		if err := runListCmd(nil, nil); err != nil {
			t.Fatalf("runListCmd() error: %v", err)
		}
	})

	if !strings.Contains(output, "alpha") {
		t.Errorf("expected output to contain %q, got: %q", "alpha", output)
	}
	if !strings.Contains(output, "beta") {
		t.Errorf("expected output to contain %q, got: %q", "beta", output)
	}
}

func TestRunListCmd_EmptyTemplatesDir(t *testing.T) {
	setupFakeRepo(t) // creates empty templates dir

	output := captureStdout(t, func() {
		if err := runListCmd(nil, nil); err != nil {
			t.Fatalf("runListCmd() error: %v", err)
		}
	})

	if output != "" {
		t.Errorf("expected no output for empty templates dir, got: %q", output)
	}
}

func TestRunViewCmd_PrintsContent(t *testing.T) {
	templatesDir := setupFakeRepo(t)
	wantContent := "# Simple Template\n\nHello!"
	if err := os.WriteFile(filepath.Join(templatesDir, "simple.md"), []byte(wantContent), 0644); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		if err := runViewCmd(nil, []string{"simple"}); err != nil {
			t.Fatalf("runViewCmd() error: %v", err)
		}
	})

	if output != wantContent {
		t.Errorf("expected %q, got %q", wantContent, output)
	}
}

func TestRunViewCmd_WithExtension(t *testing.T) {
	templatesDir := setupFakeRepo(t)
	wantContent := "# Reviewer First\n\nContent here."
	if err := os.WriteFile(filepath.Join(templatesDir, "reviewer-first.md"), []byte(wantContent), 0644); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		if err := runViewCmd(nil, []string{"reviewer-first.md"}); err != nil {
			t.Fatalf("runViewCmd() error: %v", err)
		}
	})

	if output != wantContent {
		t.Errorf("expected %q, got %q", wantContent, output)
	}
}

func TestRunViewCmd_TemplateNotFound(t *testing.T) {
	setupFakeRepo(t)

	err := runViewCmd(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got: %v", err)
	}
}

func TestResolveTemplate_NoneSet(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("GH_QPR_DEFAULT_TEMPLATE", "")

	_, err := resolveTemplate("")
	if err == nil {
		t.Fatal("expected error when no template is set, got nil")
	}
	if !strings.Contains(err.Error(), "no template specified") {
		t.Errorf("unexpected error message: %v", err)
	}
}
