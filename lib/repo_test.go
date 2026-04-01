package lib

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestListTemplates(t *testing.T) {
	t.Run("returns nil for missing templates dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		rc := &RepoCache{Path: tmpDir}
		names, err := rc.ListTemplates()
		if err != nil {
			t.Fatalf("ListTemplates() unexpected error: %v", err)
		}
		if len(names) != 0 {
			t.Errorf("expected empty slice, got %v", names)
		}
	})

	t.Run("returns template names without extension", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"alpha.md", "beta.md", "notes.txt"} {
			if err := os.WriteFile(filepath.Join(templatesDir, name), []byte(""), 0644); err != nil {
				t.Fatal(err)
			}
		}

		rc := &RepoCache{Path: tmpDir}
		names, err := rc.ListTemplates()
		if err != nil {
			t.Fatalf("ListTemplates() unexpected error: %v", err)
		}
		if len(names) != 2 {
			t.Fatalf("expected 2 templates, got %d: %v", len(names), names)
		}
		if names[0] != "alpha" || names[1] != "beta" {
			t.Errorf("unexpected template names: %v", names)
		}
	})

	t.Run("skips subdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatesDir := filepath.Join(tmpDir, "templates")
		if err := os.MkdirAll(filepath.Join(templatesDir, "subdir"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(templatesDir, "foo.md"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		rc := &RepoCache{Path: tmpDir}
		names, err := rc.ListTemplates()
		if err != nil {
			t.Fatalf("ListTemplates() unexpected error: %v", err)
		}
		if len(names) != 1 || names[0] != "foo" {
			t.Errorf("expected [foo], got %v", names)
		}
	})
}

func TestUpdate(t *testing.T) {
	oldSyncCommand := syncCommand
	oldCloneCommand := cloneCommand
	defer func() {
		syncCommand = oldSyncCommand
		cloneCommand = oldCloneCommand
	}()

	t.Run("updates existing cloned repository successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		repoPath := filepath.Join(tmpDir, "gh-qpr-repo")

		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create test repo directory: %v", err)
		}

		var syncCalled bool
		syncCommand = func(owner, repoName, path string) *exec.Cmd {
			syncCalled = true
			if path != repoPath {
				t.Errorf("syncCommand called with unexpected path: got %s, want %s", path, repoPath)
			}
			return exec.Command("true")
		}
		cloneCommand = func(owner, repoName, path string) *exec.Cmd {
			t.Fatalf("cloneCommand should not be called if repo exists")
			return exec.Command("false")
		}

		rc := &RepoCache{
			Owner:    "testowner",
			RepoName: "testrepo",
			Path:     repoPath,
		}

		if err := rc.Update(); err != nil {
			t.Errorf("Update() failed unexpectedly: %v", err)
		}
		if !syncCalled {
			t.Error("syncCommand was not called")
		}
	})

	t.Run("clones and then updates if repository does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		repoPath := filepath.Join(tmpDir, "gh-qpr-repo-uncloned")

		var cloneCalled, syncCalled bool

		cloneCommand = func(owner, repoName, path string) *exec.Cmd {
			cloneCalled = true
			return exec.Command("true")
		}
		syncCommand = func(owner, repoName, path string) *exec.Cmd {
			syncCalled = true
			return exec.Command("true")
		}

		rc := &RepoCache{
			Owner:    "testowner",
			RepoName: "testrepo",
			Path:     repoPath,
		}

		if err := rc.Update(); err != nil {
			t.Fatalf("Update() failed unexpectedly: %v", err)
		}
		if !cloneCalled {
			t.Error("cloneCommand was not called when repository did not exist")
		}
		if !syncCalled {
			t.Error("syncCommand was not called after cloning")
		}
	})

	t.Run("returns error if gh repo sync fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		repoPath := filepath.Join(tmpDir, "gh-qpr-repo-sync-fail")

		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create test repo directory: %v", err)
		}

		var syncCalled bool
		syncCommand = func(owner, repoName, path string) *exec.Cmd {
			syncCalled = true
			return exec.Command("false")
		}
		cloneCommand = func(owner, repoName, path string) *exec.Cmd {
			return exec.Command("true")
		}

		rc := &RepoCache{
			Owner:    "testowner",
			RepoName: "testrepo",
			Path:     repoPath,
		}

		err := rc.Update()
		if err == nil {
			t.Error("Update() did not return an error when gh repo sync failed")
		}
		if !strings.Contains(err.Error(), "failed to update repository") {
			t.Errorf("unexpected error message: %v", err)
		}
		if !syncCalled {
			t.Error("syncCommand was not called")
		}
	})
}
