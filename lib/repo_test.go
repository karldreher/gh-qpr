package lib

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetRepoFromEnv(t *testing.T) {
	// Test case 1: GH_QPR_REPO not set
	owner, repo := GetRepoFromEnv()
	if owner != "karldreher" || repo != "gh-qpr" {
		t.Errorf("Expected karldreher/gh-qpr, but got %s/%s", owner, repo)
	}

	// Test case 2: GH_QPR_REPO is set
	os.Setenv("GH_QPR_REPO", "owner/repo")
	owner, repo = GetRepoFromEnv()
	if owner != "owner" || repo != "repo" {
		t.Errorf("Expected owner/repo, but got %s/%s", owner, repo)
	}
	os.Unsetenv("GH_QPR_REPO")
}

func TestTemplatePath(t *testing.T) {
	rc := &RepoCache{
		Owner:    "owner",
		RepoName: "repo",
		Path:     "/tmp/gh-qpr/repo",
	}

	// Test case 1: Template name with extension
	templateName := "template.md"
	expectedPath := filepath.Join(rc.Path, "templates", templateName)
	path := rc.TemplatePath(templateName)
	if path != expectedPath {
		t.Errorf("Expected %s, but got %s", expectedPath, path)
	}

	// Test case 2: Template name without extension
	templateName = "template"
	expectedPath = filepath.Join(rc.Path, "templates", templateName+".md")
	path = rc.TemplatePath(templateName)
	if path != expectedPath {
		t.Errorf("Expected %s, but got %s", expectedPath, path)
	}
}

func TestEnsureCloned(t *testing.T) {
	// Replace cloneCommand to avoid actual cloning
	oldCloneCommand := cloneCommand
	defer func() { cloneCommand = oldCloneCommand }()
	cloneCommand = func(owner, repoName, path string) *exec.Cmd {
		// The command can be anything that is likely to succeed without side effects.
		// On Unix-like systems, `true` is a command that does nothing and exits with status 0.
		return exec.Command("true")
	}

	t.Run("creates directory if it does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		repoPath := filepath.Join(tmpDir, "a", "b", "c")

		// a/b should be created by EnsureCloned
		parentDir := filepath.Dir(repoPath)

		rc := &RepoCache{
			Owner:    "owner",
			RepoName: "repo",
			Path:     repoPath,
		}

		// Pre-condition: parent directory should not exist
		if _, err := os.Stat(parentDir); !os.IsNotExist(err) {
			t.Fatalf("parent directory %s already exists before test", parentDir)
		}

		if err := rc.EnsureCloned(); err != nil {
			t.Fatalf("EnsureCloned() failed: %v", err)
		}

		// Post-condition: Check if parent directory was created
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			t.Errorf("EnsureCloned() did not create parent directory %s", parentDir)
		}
	})

	t.Run("does nothing if directory already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		repoPath := filepath.Join(tmpDir, "a")

		// Create the directory beforehand
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		rc := &RepoCache{
			Owner:    "owner",
			RepoName: "repo",
			Path:     repoPath,
		}

		if err := rc.EnsureCloned(); err != nil {
			t.Fatalf("EnsureCloned() failed: %v", err)
		}
	})
}