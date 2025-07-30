package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Flags
	templateFlag := flag.String("template", "", "template file name")
	editFlag := flag.String("edit", "", "Edit existing PR, rather than create it.  Overwrites PR description.")
	flag.Parse()

	if *templateFlag == "" {
		fmt.Println("Please provide a template name using the --template flag.")
		return
	}
	// TODO : This is not a path to the file in this folder.
	// It is a path to the template relative to the "qpr-repo";
	// The default of which is karldreher/gh-qpr,
	// but can be overridden with the GH_QPR_REPO environment variable.
	// This implies a fairly broad set of changes.
	templatePath := filepath.Join("templates", *templateFlag)
	if filepath.Ext(templatePath) == "" {
		// Always assume this is md,
		// so users can either provide this or not.
		templatePath += ".md"
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("Error reading template file: %v\n", err)
		return
	}
	// The subcommand that is passed to GH.
	var subcommand string
	if *editFlag != "" {
		subcommand = "edit"
	} else {
		subcommand = "create"
	}
	cmd := exec.Command("gh", "pr", subcommand, "--body", string(content))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error creating pull request: %v\n", err)
	}
}
