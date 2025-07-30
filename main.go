package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	templateName := flag.String("template", "", "template file name")
	flag.Parse()

	if *templateName == "" {
		fmt.Println("Please provide a template name using the --template flag.")
		return
	}

	templatePath := filepath.Join("templates", *templateName)
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

	cmd := exec.Command("gh", "pr", "create", "--body", string(content))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error creating pull request: %v\n", err)
	}
}
