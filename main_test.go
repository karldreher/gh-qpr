package main

import (
	"encoding/json"
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
