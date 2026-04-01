package lib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() failed: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join(".gh-qpr", "config.json")) {
		t.Errorf("unexpected config path: %s", path)
	}
	if !strings.HasPrefix(path, tmpDir) {
		t.Errorf("config path %s is not under HOME %s", path, tmpDir)
	}
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error for missing file: %v", err)
	}
	if cfg.DefaultTemplate != "" {
		t.Errorf("expected empty DefaultTemplate, got %q", cfg.DefaultTemplate)
	}
}

func TestLoadConfig_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".gh-qpr")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(Config{DefaultTemplate: "simple"})
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}
	if cfg.DefaultTemplate != "simple" {
		t.Errorf("expected DefaultTemplate %q, got %q", "simple", cfg.DefaultTemplate)
	}
}

func TestLoadConfig_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".gh-qpr")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte("{bad json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() expected error for malformed JSON, got nil")
	}
}

func TestLoadConfig_WrongFieldType(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".gh-qpr")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	// default_template must be a string; passing a number violates the schema.
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"default_template": 123}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() expected error for wrong field type, got nil")
	}
	if !strings.Contains(err.Error(), "parsing config file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := Config{DefaultTemplate: "reviewer-first"}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after SaveConfig() failed: %v", err)
	}
	if loaded.DefaultTemplate != cfg.DefaultTemplate {
		t.Errorf("round-trip mismatch: got %q, want %q", loaded.DefaultTemplate, cfg.DefaultTemplate)
	}
}

func TestSaveConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Confirm ~/.gh-qpr does not exist yet
	configDir := filepath.Join(tmpDir, ".gh-qpr")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("config directory should not exist before SaveConfig")
	}

	if err := SaveConfig(Config{DefaultTemplate: "simple"}); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("SaveConfig() did not create the config directory")
	}
}
