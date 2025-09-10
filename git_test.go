package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, func())
		expectError bool
		errorType   error
	}{
		{
			name: "finds git root in current directory",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				return tempDir, func() {}
			},
			expectError: false,
		},
		{
			name: "finds git root in parent directory", 
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				subDir := filepath.Join(tempDir, "subdir")
				if err := os.Mkdir(subDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				return subDir, func() {}
			},
			expectError: false,
		},
		{
			name: "finds git root with nested subdirectories",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				deepDir := filepath.Join(tempDir, "a", "b", "c")
				if err := os.MkdirAll(deepDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				return deepDir, func() {}
			},
			expectError: false,
		},
		{
			name: "handles git worktree (.git file)",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				gitFile := filepath.Join(tempDir, ".git")
				if err := os.WriteFile(gitFile, []byte("gitdir: /path/to/git"), 0644); err != nil {
					t.Fatal(err)
				}
				return tempDir, func() {}
			},
			expectError: false,
		},
		{
			name: "returns error when no git repository found",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				return tempDir, func() {}
			},
			expectError: true,
			errorType:   ErrNotGitRepo,
		},
		{
			name: "handles file path input",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				testFile := filepath.Join(tempDir, "test.txt")
				if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
				
				return testFile, func() {}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			result, err := FindGitRoot(testPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify that the result contains a .git directory or file
			gitPath := filepath.Join(result, ".git")
			if _, err := os.Stat(gitPath); err != nil {
				t.Errorf("result %s does not contain .git: %v", result, err)
			}
		})
	}
}

func TestFindGitRootRealRepository(t *testing.T) {
	// Test with the actual git repository of this project
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	root, err := FindGitRoot(wd)
	if err != nil {
		t.Skipf("skipping test in non-git environment: %v", err)
		return
	}

	// Verify the result is a valid git root
	gitPath := filepath.Join(root, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		t.Errorf("result %s does not contain .git: %v", root, err)
	}

	// Verify it returns the same result when called with a file in the repo
	thisFile := filepath.Join(wd, "git_test.go")
	root2, err := FindGitRoot(thisFile)
	if err != nil {
		t.Errorf("failed to find git root from file path: %v", err)
	}

	if root != root2 {
		t.Errorf("git root differs when called with directory vs file: %s vs %s", root, root2)
	}
}