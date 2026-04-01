package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureClonedIntegration(t *testing.T) {
	if os.Getenv("GH_TOKEN") == "" {
		t.Skip("GH_TOKEN not set, skipping integration test")
	}

	tmpDir := t.TempDir()
	rc := &RepoCache{
		Owner:    "karldreher",
		RepoName: "gh-qpr",
		Path:     filepath.Join(tmpDir, "gh-qpr"),
	}

	if err := rc.EnsureCloned(); err != nil {
		t.Fatalf("EnsureCloned() failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(rc.Path, ".git")); err != nil {
		t.Errorf("expected .git directory in cloned repo, got: %v", err)
	}
}

func TestUpdateIntegration(t *testing.T) {
	if os.Getenv("GH_TOKEN") == "" {
		t.Skip("GH_TOKEN not set, skipping integration test")
	}

	tmpDir := t.TempDir()
	rc := &RepoCache{
		Owner:    "karldreher",
		RepoName: "gh-qpr",
		Path:     filepath.Join(tmpDir, "gh-qpr"),
	}

	if err := rc.EnsureCloned(); err != nil {
		t.Fatalf("EnsureCloned() setup failed: %v", err)
	}

	if err := rc.Update(); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
}
