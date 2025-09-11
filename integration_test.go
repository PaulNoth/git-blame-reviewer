package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMainIntegration(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "test-git-review-blame", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("test-git-review-blame")

	tests := []struct {
		name           string
		args           []string
		env            map[string]string
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
			name:           "no tokens provided",
			args:           []string{"main.go"},
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
			cmd := exec.Command("./test-git-review-blame", tt.args...)
			
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
	buildCmd := exec.Command("go", "build", "-o", "test-git-review-blame", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("test-git-review-blame")

	// Test with valid git repository but dummy token (will fail at API stage)
	cmd := exec.Command("./test-git-review-blame", "-porcelain", "-show-email", "main.go") 
	cmd.Env = append(os.Environ(), "GITHUB_TOKEN=dummy-token")
	
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Should fail but not due to flag parsing
	if err == nil {
		t.Error("Expected command to fail (no real API token)")
	}

	// Should not contain flag parsing errors
	if strings.Contains(outputStr, "flag provided but not defined") {
		t.Errorf("Flag parsing failed: %s", outputStr)
	}
}