package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		filePath = flag.String("file", "", "Path to the file to blame")
		token    = flag.String("token", "", "GitHub token (or set GITHUB_TOKEN env var)")
		verbose  = flag.Bool("verbose", false, "Enable verbose output")
		help     = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *filePath == "" {
		fmt.Fprintf(os.Stderr, "Error: file path is required\n")
		fmt.Fprintf(os.Stderr, "Use -help for usage information\n")
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Processing file: %s\n", *filePath)
	}

	// TODO: Implement the main logic
	fmt.Printf("git-review-blame: Processing %s\n", *filePath)
	fmt.Printf("GitHub token provided: %t\n", *token != "")
}

func showHelp() {
	fmt.Printf(`git-review-blame - Show GitHub PR approvers for each line instead of commit authors

Usage:
  git-review-blame -file <path> [-token <github-token>] [-verbose]

Options:
  -file string
    	Path to the file to blame
  -token string
    	GitHub token (or set GITHUB_TOKEN env var)
  -verbose
    	Enable verbose output
  -help
    	Show this help message

Examples:
  git-review-blame -file src/main.go
  git-review-blame -file src/main.go -token ghp_xxxx
  GITHUB_TOKEN=ghp_xxxx git-review-blame -file src/main.go

Environment Variables:
  GITHUB_TOKEN - GitHub personal access token
`)
}