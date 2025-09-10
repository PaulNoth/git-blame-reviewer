package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		lineNumber = flag.String("L", "", "Annotate only the given line range")
		porcelain  = flag.Bool("porcelain", false, "Show in a format designed for machine consumption")
		showRoot   = flag.Bool("root", false, "Do not treat root commits as boundaries")
		showEmail  = flag.Bool("show-email", false, "Show author email instead of author name")
		help       = flag.Bool("help", false, "Show help message")
	)
	
	// Parse flags first
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Get the file path from remaining arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "fatal: no file specified\n")
		os.Exit(1)
	}

	filePath := args[0]

	// Get GitHub token from environment if not provided via flag
	token := os.Getenv("GITHUB_TOKEN")

	// TODO: Implement the main logic
	fmt.Printf("git-review-blame: Processing %s\n", filePath)
	fmt.Printf("Line range: %s\n", *lineNumber)
	fmt.Printf("Porcelain: %t\n", *porcelain)
	fmt.Printf("Show root: %t\n", *showRoot)
	fmt.Printf("Show email: %t\n", *showEmail)
	fmt.Printf("GitHub token provided: %t\n", token != "")
}

func showHelp() {
	fmt.Printf(`git-review-blame - Show GitHub PR approvers for each line instead of commit authors

Usage:
  git-review-blame [<options>] [<rev-opts>] [<rev>] [--] <file>

Options:
  -L <start>,<end>    Show only lines in given range
  -porcelain          Show in a format designed for machine consumption  
  -root               Do not treat root commits as boundaries
  -show-email         Show author email instead of author name
  -help               Show this help message

Environment Variables:
  GITHUB_TOKEN - GitHub personal access token (required)

Examples:
  git-review-blame src/main.go
  git-review-blame -L 10,20 src/main.go  
  git-review-blame -porcelain src/main.go
`)
}