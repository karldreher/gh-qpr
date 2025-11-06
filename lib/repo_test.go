package lib

import (
	"os"
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
