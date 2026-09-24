package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitBlameFirstLineAuthor returns the author name git blame reports for the
// first line of the given file in the current repository. Used instead of
// hardcoding a specific historical author name, since that would only hold
// in a full local clone: CI's actions/checkout defaults to a shallow clone
// (fetch-depth: 1), where git blame can only see the single fetched commit
// and therefore attributes every line to that commit's author instead of
// the true historical author.
func gitBlameFirstLineAuthor(t *testing.T, file string) string {
	t.Helper()

	out, err := exec.CommandContext(context.Background(), "git", "blame", "--porcelain", "-L", "1,1", file).Output()
	if err != nil {
		t.Fatalf("failed to run git blame on %s: %v", file, err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if author, ok := strings.CutPrefix(line, "author "); ok {
			return author
		}
	}

	t.Fatalf("could not find an author line in git blame output for %s", file)
	return ""
}

// newIsolatedGitRepoWithoutOrigin creates a fresh git repository with no
// "origin" remote configured, so that ExtractRepoInfo cannot determine the
// repository type. Used to test the "could not determine repo type" error
// path in isolation, regardless of which repository the test suite itself
// happens to run in.
func newIsolatedGitRepoWithoutOrigin(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	initCmd := exec.CommandContext(context.Background(), "git", "init")
	initCmd.Dir = dir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init isolated git repo: %v", err)
	}

	return dir
}

func TestMainIntegration(t *testing.T) {
	// Build the binary first
	binPath, err := filepath.Abs("test-git-review-blame")
	if err != nil {
		t.Fatalf("Failed to resolve binary path: %v", err)
	}
	buildCmd := exec.CommandContext(context.Background(), "go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binPath)

	tests := []struct {
		name           string
		args           []string
		env            map[string]string
		dir            string
		expectExitCode int
		expectOutput   string
		expectError    string
	}{
		{
			name:           "help flag",
			args:           []string{"-help"},
			expectExitCode: 0,
			expectOutput:   "git-review-blame - Show GitHub/GitLab PR/MR approvers",
		},
		{
			name:           "no file specified",
			args:           []string{},
			expectExitCode: 1,
			expectError:    "Error: Please specify a file to analyze",
		},
		{
			name:           "no tokens provided, repo type undetermined",
			args:           []string{"somefile.go"},
			dir:            newIsolatedGitRepoWithoutOrigin(t),
			expectExitCode: 1,
			expectError:    "Error: could not determine if this is a GitHub or GitLab repository",
		},
		{
			name: "GitHub repo with GitHub token",
			args: []string{"/tmp/nonexistent.go"},
			env: map[string]string{
				"GITHUB_TOKEN": "dummy-token",
			},
			expectExitCode: 1,
			expectError:    "Error: this directory is not part of a Git repository",
		},
		{
			name: "GitLab repo with GitLab token",
			args: []string{"/tmp/nonexistent.go"},
			env: map[string]string{
				"GITLAB_TOKEN": "dummy-token",
			},
			expectExitCode: 1,
			expectError:    "Error: this directory is not part of a Git repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// binPath and tt.args are both fixed, compile-time test data
			// (the freshly built test binary and this file's own table
			// above), not external/user input.
			cmd := exec.CommandContext(context.Background(), binPath, tt.args...) //nolint:gosec // G204: test-controlled binary path and args
			if tt.dir != "" {
				cmd.Dir = tt.dir
			}

			// Set environment variables
			if tt.env != nil {
				env := os.Environ()
				for key, value := range tt.env {
					env = append(env, key+"="+value)
				}
				cmd.Env = env
			}

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("Failed to run command: %v", err)
				}
			}

			if exitCode != tt.expectExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectExitCode, exitCode)
			}

			// Check expected output
			if tt.expectOutput != "" {
				if !strings.Contains(outputStr, tt.expectOutput) {
					t.Errorf("Expected output to contain %q, got:\n%s", tt.expectOutput, outputStr)
				}
			}

			// Check expected error
			if tt.expectError != "" {
				if !strings.Contains(outputStr, tt.expectError) {
					t.Errorf("Expected error to contain %q, got:\n%s", tt.expectError, outputStr)
				}
			}
		})
	}
}

func TestMainFlags(t *testing.T) {
	// Test that flags are parsed correctly
	binPath, err := filepath.Abs("test-git-review-blame-flags")
	if err != nil {
		t.Fatalf("Failed to resolve binary path: %v", err)
	}
	buildCmd := exec.CommandContext(context.Background(), "go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binPath)

	// Determine the expected fallback author dynamically from git blame
	// itself, rather than hardcoding a specific name (see
	// gitBlameFirstLineAuthor's comment for why).
	expectedAuthor := gitBlameFirstLineAuthor(t, "main.go")

	// With a valid git repository but an invalid/dummy API token, PR approval
	// lookups fail per-commit. The tool is designed to gracefully fall back
	// to the original commit's author/date in that case (see README: "Falls
	// back to original commit info if no PR/MR/approval found"), rather than
	// aborting the whole command. So this should succeed (exit 0) and still
	// produce valid blame output built from the original commit metadata.
	cmd := exec.CommandContext(context.Background(), binPath, "-porcelain", "-show-email", "main.go")
	cmd.Env = append(os.Environ(), "GITHUB_TOKEN=dummy-token")

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		t.Errorf("Expected command to succeed via fallback to original commit info, got error: %v\nOutput:\n%s", err, outputStr)
	}

	// Should not contain flag parsing errors
	if strings.Contains(outputStr, "flag provided but not defined") {
		t.Errorf("Flag parsing failed: %s", outputStr)
	}

	// Should still produce valid porcelain blame output, falling back to the
	// original commit author since the dummy token can't authenticate.
	if !strings.Contains(outputStr, "author "+expectedAuthor) {
		t.Errorf("Expected fallback blame output with author %q, got:\n%s", expectedAuthor, outputStr)
	}
}

func TestHelpDocumentsBoundaryFlag(t *testing.T) {
	binPath, err := filepath.Abs("test-git-review-blame-help-b")
	if err != nil {
		t.Fatalf("Failed to resolve binary path: %v", err)
	}
	buildCmd := exec.CommandContext(context.Background(), "go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binPath)

	cmd := exec.CommandContext(context.Background(), binPath, "-help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run -help: %v\nOutput:\n%s", err, output)
	}

	if !strings.Contains(string(output), "-b ") {
		t.Errorf("Expected help output to document the -b flag, got:\n%s", output)
	}
}

func TestMainBoundaryFlag(t *testing.T) {
	// go.mod's "go 1.25.1" line traces back to this repository's initial
	// (root) commit, making it a boundary line for `git blame`.
	binPath, err := filepath.Abs("test-git-review-blame-boundary")
	if err != nil {
		t.Fatalf("Failed to resolve binary path: %v", err)
	}
	buildCmd := exec.CommandContext(context.Background(), "go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binPath)

	runWithEnv := func(args ...string) string {
		cmd := exec.CommandContext(context.Background(), binPath, args...)
		// Set both tokens: whether this checkout's remote resolves to GitHub
		// or GitLab depends on the local git config's URL rewrites (e.g. SSH
		// host aliases), which varies across machines/CI.
		cmd.Env = append(os.Environ(), "GITHUB_TOKEN=dummy-token", "GITLAB_TOKEN=dummy-token")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\nOutput:\n%s", args, err, output)
		}
		return string(output)
	}

	humanNoFlag := runWithEnv("-L", "3,3", "go.mod")
	humanWithFlag := runWithEnv("-b", "-L", "3,3", "go.mod")

	if humanNoFlag == humanWithFlag {
		t.Errorf("expected -b to change human-readable output for a boundary line, got identical output:\n%s", humanNoFlag)
	}
	identifierField := strings.TrimSpace(strings.SplitN(humanWithFlag, "(", 2)[0])
	if identifierField != "" {
		t.Errorf("expected blank identifier column with -b on a boundary line, got %q in:\n%s", identifierField, humanWithFlag)
	}

	porcelainNoFlag := runWithEnv("-porcelain", "-L", "3,3", "go.mod")
	porcelainWithFlag := runWithEnv("-b", "-porcelain", "-L", "3,3", "go.mod")

	if porcelainNoFlag != porcelainWithFlag {
		t.Errorf("expected -b to have no effect on -porcelain output, got:\nwithout=%q\nwith=%q", porcelainNoFlag, porcelainWithFlag)
	}
}
