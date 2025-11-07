package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoCache manages the local cache of the template repository.
type RepoCache struct {
	Owner    string
	RepoName string
	Path     string
}

// NewRepoCache creates a new RepoCache instance.
func NewRepoCache(owner, repoName string) (*RepoCache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	cachePath := filepath.Join(home, ".gh-qpr", repoName)
	return &RepoCache{
			Owner:    owner,
			RepoName: repoName,
			Path:     cachePath,
		},
		nil
}

var cloneCommand = func(owner, repoName, path string) *exec.Cmd {
	return exec.Command("gh", "repo", "clone", fmt.Sprintf("%s/%s", owner, repoName), path)
}

// EnsureCloned ensures that the repository is cloned to the cache directory.
func (rc *RepoCache) EnsureCloned() error {
	if _, err := os.Stat(rc.Path); os.IsNotExist(err) {
		fmt.Printf("Cloning %s/%s to %s...\n", rc.Owner, rc.RepoName, rc.Path)
		parentDir := filepath.Dir(rc.Path)
		// Directories must be created with 0755
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory %s: %w", parentDir, err)
		}
		cmd := cloneCommand(rc.Owner, rc.RepoName, rc.Path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

// TemplatePath returns the full path to a template file in the cached repository.
func (rc *RepoCache) TemplatePath(templateName string) string {
	templatePath := filepath.Join(rc.Path, "templates", templateName)
	// A template may be referenced with or without the .md extension
	if filepath.Ext(templatePath) == "" {
		templatePath += ".md"
	}
	return templatePath
}

func GetRepoFromEnv() (string, string) {
	repoEnv := os.Getenv("GH_QPR_REPO")
	if repoEnv == "" {
		return "karldreher", "gh-qpr"
	}
	parts := strings.Split(repoEnv, "/")
	if len(parts) != 2 {
		fmt.Println("Invalid format for GH_QPR_REPO. Please use 'owner/repo'.")
		os.Exit(1)
	}
	return parts[0], parts[1]
}
