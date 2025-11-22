package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/karldreher/gh-qpr/lib"
	"github.com/spf13/cobra"
)

func runTemplatePR(action string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		templateName, _ := cmd.Flags().GetString("template")
		if templateName == "" {
			return fmt.Errorf("--template flag is required")
		}

		repoOwner, repoName := lib.GetRepoFromEnv()
		repoCache, err := lib.NewRepoCache(repoOwner, repoName)
		if err != nil {
			return fmt.Errorf("error creating repo cache: %v", err)
		}

		if err := repoCache.EnsureCloned(); err != nil {
			return fmt.Errorf("error cloning repository: %v", err)
		}

		templatePath := repoCache.TemplatePath(templateName)
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("error reading template file: %v", err)
		}

		cmdArgs := []string{"pr", action, "--body", string(content)}
		if action == "create" {
			title, _ := cmd.Flags().GetString("title")
			if title == "" {
				// PR titles can always be changed.
				// This provides a safe default.
				// If the user wants a specific title, they can use the --title flag.
				title = "QPR"
			}
			cmdArgs = append(cmdArgs, "--title", title)
		}

		cmdExec := exec.Command("gh", cmdArgs...)
		cmdExec.Stdout = os.Stdout

		var stderrBuf bytes.Buffer
		cmdExec.Stderr = &stderrBuf

		err = cmdExec.Run()

		if err != nil {
			stderrOutput := stderrBuf.String()
			if strings.Contains(stderrOutput, "GraphQL: Projects (classic) is being deprecated") {
				return fmt.Errorf(
					"error running gh pr %s: The `gh` CLI encountered a deprecated GitHub Projects API. " +
						"This is likely due to an outdated `gh` CLI version. " +
						"Please update your GitHub CLI to the latest version. " +
						"Original error: %w", err)
			}
			return fmt.Errorf("error running gh pr %s: %w", action, err)
		}
		return nil
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gh-qpr",
		Short: "GitHub PR helper for templates",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new pull request using a template",
		RunE:  runTemplatePR("create"),
	}
	createCmd.Flags().StringP("template", "t", "", "template file name (required)")
	createCmd.Flags().StringP("title", "T", "", "pull request title (default: QPR)")
	createCmd.MarkFlagRequired("template")

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an existing pull request using a template.  WARNING: Overwrites existing description.",
		RunE:  runTemplatePR("edit"),
	}
	editCmd.Flags().StringP("template", "t", "", "template file name (required)")
	editCmd.MarkFlagRequired("template")

	rootCmd.AddCommand(createCmd, editCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
