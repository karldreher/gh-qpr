package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/karldreher/gh-qpr/lib"
	"github.com/spf13/cobra"
)

// resolveTemplate determines the effective template name, checking (in order):
// the explicit flag value, GH_QPR_DEFAULT_TEMPLATE env var, and the config file.
func resolveTemplate(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if v := os.Getenv("GH_QPR_DEFAULT_TEMPLATE"); v != "" {
		return v, nil
	}
	cfg, err := lib.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}
	if cfg.DefaultTemplate != "" {
		return cfg.DefaultTemplate, nil
	}
	return "", fmt.Errorf("no template specified: use --template, set GH_QPR_DEFAULT_TEMPLATE, or run 'gh qpr default'")
}

// runDefaultCmd is the handler for `gh qpr default`. It lists available templates
// from the local cache and prompts the user to select one, then persists the
// choice to ~/.gh-qpr/config.json.
func runDefaultCmd(cmd *cobra.Command, args []string) error {
	repoOwner, repoName := lib.GetRepoFromEnv()
	repoCache, err := lib.NewRepoCache(repoOwner, repoName)
	if err != nil {
		return fmt.Errorf("error creating repo cache: %v", err)
	}
	if err := repoCache.EnsureCloned(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}
	templates, err := repoCache.ListTemplates()
	if err != nil {
		return fmt.Errorf("error listing templates: %v", err)
	}
	if len(templates) == 0 {
		return fmt.Errorf("no templates found in %s", repoCache.Path)
	}
	fmt.Println("Available templates:")
	for i, name := range templates {
		fmt.Printf("  %d) %s\n", i+1, name)
	}
	fmt.Printf("Select template [1-%d]: ", len(templates))
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	idx, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || idx < 1 || idx > len(templates) {
		return fmt.Errorf("invalid selection: %q", strings.TrimSpace(line))
	}
	chosen := templates[idx-1]
	cfg, err := lib.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	cfg.DefaultTemplate = chosen
	if err := lib.SaveConfig(cfg); err != nil {
		return fmt.Errorf("error saving config: %v", err)
	}
	fmt.Printf("Default template set to: %s\n", chosen)
	return nil
}

// runTemplatePR returns a Cobra RunE handler for the given PR action ("create" or "edit").
// It resolves the template via resolveTemplate, reads the template body from the local
// cache, and invokes `gh pr <action>` with that body.
func runTemplatePR(action string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		templateFlagValue, _ := cmd.Flags().GetString("template")
		templateName, err := resolveTemplate(templateFlagValue)
		if err != nil {
			return err
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
		cmdExec.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

		err = cmdExec.Run()

		if err != nil {

			stderrOutput := stderrBuf.String()

			if strings.Contains(stderrOutput, "GraphQL: Projects (classic) is being deprecated") {

				return fmt.Errorf(

					"error running gh pr %s: The `gh` CLI encountered a deprecated GitHub Projects API. "+

						"This is likely due to an outdated `gh` CLI version. "+

						"Please update your GitHub CLI to the latest version. "+

						"Original error: %v", action, err)

			}

			return fmt.Errorf("error running gh pr %s: %w", action, err)

		}

		return nil

	}

}

// runListCmd is the handler for `gh qpr list`. It prints the names of all
// available templates in the local cache.
func runListCmd(cmd *cobra.Command, args []string) error {
	repoOwner, repoName := lib.GetRepoFromEnv()
	repoCache, err := lib.NewRepoCache(repoOwner, repoName)
	if err != nil {
		return fmt.Errorf("error creating repo cache: %v", err)
	}
	if err := repoCache.EnsureCloned(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}
	templates, err := repoCache.ListTemplates()
	if err != nil {
		return fmt.Errorf("error listing templates: %v", err)
	}
	for _, name := range templates {
		fmt.Println(name)
	}
	return nil
}

// runViewCmd is the handler for `gh qpr view <template>`. It prints the content
// of the named template from the local cache.
func runViewCmd(cmd *cobra.Command, args []string) error {
	repoOwner, repoName := lib.GetRepoFromEnv()
	repoCache, err := lib.NewRepoCache(repoOwner, repoName)
	if err != nil {
		return fmt.Errorf("error creating repo cache: %v", err)
	}
	if err := repoCache.EnsureCloned(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}
	templatePath := repoCache.TemplatePath(args[0])
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("template %q not found: %v", args[0], err)
	}
	fmt.Print(string(content))
	return nil
}

// runUpdateRepo is the handler for `gh qpr update`. It syncs the local template
// repo cache with the latest changes from the remote.
func runUpdateRepo(cmd *cobra.Command, args []string) error {

	repoOwner, repoName := lib.GetRepoFromEnv()

	repoCache, err := lib.NewRepoCache(repoOwner, repoName)

	if err != nil {

		return fmt.Errorf("error creating repo cache: %v", err)

	}

	if err := repoCache.Update(); err != nil {

		return fmt.Errorf("error updating repository cache: %v", err)

	}

	fmt.Println("Repository cache updated successfully.")

	return nil

}

func main() {

	var rootCmd = &cobra.Command{

		Use: "gh-qpr",

		Short: "GitHub PR helper for templates",
	}

	createCmd := &cobra.Command{

		Use: "create",

		Short: "Create a new pull request using a template",

		RunE: runTemplatePR("create"),
	}

	createCmd.Flags().StringP("template", "t", "", "template file name (uses default if not set)")

	createCmd.Flags().StringP("title", "T", "", "pull request title (default: QPR)")

	editCmd := &cobra.Command{

		Use: "edit",

		Short: "Edit an existing pull request using a template.  WARNING: Overwrites existing description.",

		RunE: runTemplatePR("edit"),
	}

	editCmd.Flags().StringP("template", "t", "", "template file name (uses default if not set)")

	updateCmd := &cobra.Command{

		Use: "update",

		Short: "Update the local repository cache with the latest changes from remote.",

		RunE: runUpdateRepo,
	}

	defaultCmd := &cobra.Command{

		Use:   "default",

		Short: "Interactively select the default PR template",

		RunE: runDefaultCmd,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available PR templates",
		RunE:  runListCmd,
	}

	viewCmd := &cobra.Command{
		Use:   "view <template>",
		Short: "View the content of a PR template",
		Args:  cobra.ExactArgs(1),
		RunE:  runViewCmd,
	}

	rootCmd.AddCommand(createCmd, editCmd, updateCmd, defaultCmd, listCmd, viewCmd)

	if err := rootCmd.Execute(); err != nil {

		fmt.Fprintf(os.Stderr, "%v\n", err)

		os.Exit(1)

	}

}
