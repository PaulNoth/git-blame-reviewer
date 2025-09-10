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

	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "fatal: GITHUB_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	// Run the main logic
	if err := runGitReviewBlame(filePath, *lineNumber, *porcelain, *showEmail, token); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf(`git-review-blame - Show GitHub PR approvers for each line instead of commit authors

Usage:
  git-review-blame [<options>] [<rev-opts>] [<rev>] [--] <file>

Options:
  -L <start>,<end>    Show only lines in given range
  -porcelain          Show in a format designed for machine consumption  
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

// runGitReviewBlame executes the main logic of the application
func runGitReviewBlame(filePath, lineRange string, porcelain, showEmail bool, token string) error {
	// 1. Find git repository root
	repoRoot, err := FindGitRoot(filePath)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// 2. Extract repository information from git remote
	repoInfo, err := ExtractRepoInfo(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to extract repository info: %w", err)
	}

	// 3. Execute git blame on the file
	blameLines, err := ExecuteGitBlame(repoRoot, filePath, lineRange, porcelain)
	if err != nil {
		return fmt.Errorf("git blame failed: %w", err)
	}

	// 4. Create GitHub client
	client := NewGitHubClient(token)

	// 5. Process each blame line to get PR approval info
	var linesWithApprovals []BlameLineWithApproval
	
	// Cache to avoid duplicate API calls for same commit
	commitCache := make(map[string]*PRApprovalInfo)
	
	for _, blameLine := range blameLines {
		lineWithApproval := BlameLineWithApproval{
			BlameLine: blameLine,
		}
		
		// Check cache first
		if approvalInfo, exists := commitCache[blameLine.CommitHash]; exists {
			if approvalInfo != nil {
				lineWithApproval.PRNumber = approvalInfo.PR.Number
				if len(approvalInfo.Approvers) > 0 {
					// Use the most recent approver
					lastApprover := approvalInfo.Approvers[len(approvalInfo.Approvers)-1]
					lineWithApproval.Approver = lastApprover.User.Login
					lineWithApproval.ApproverEmail = lastApprover.User.Email
					lineWithApproval.ApprovalTime = lastApprover.SubmittedAt
				}
			}
		} else {
			// Fetch PR approval info from GitHub
			approvalInfo, err := client.GetPRApprovalInfo(repoInfo.Owner, repoInfo.Name, blameLine.CommitHash)
			if err != nil {
				// Cache the error (nil) to avoid repeated failures
				commitCache[blameLine.CommitHash] = nil
			} else {
				// Cache the result
				commitCache[blameLine.CommitHash] = approvalInfo
				
				lineWithApproval.PRNumber = approvalInfo.PR.Number
				if len(approvalInfo.Approvers) > 0 {
					// Use the most recent approver
					lastApprover := approvalInfo.Approvers[len(approvalInfo.Approvers)-1]
					lineWithApproval.Approver = lastApprover.User.Login
					lineWithApproval.ApproverEmail = lastApprover.User.Email
					lineWithApproval.ApprovalTime = lastApprover.SubmittedAt
				}
			}
		}
		
		linesWithApprovals = append(linesWithApprovals, lineWithApproval)
	}

	// 6. Format and display the output
	formatter := NewOutputFormatter(showEmail, porcelain, false)
	output := formatter.FormatOutput(linesWithApprovals)
	fmt.Print(output)

	return nil
}