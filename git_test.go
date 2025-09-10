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

func TestParseGitBlameOutput(t *testing.T) {
	// Sample porcelain output from git blame --line-porcelain
	sampleOutput := `a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0 1 1 1
author John Doe
author-mail <john.doe@example.com>
author-time 1609459200
	package main
b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1 2 2 1
author Jane Smith
author-mail <jane.smith@example.com>
author-time 1609545600
	
c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2 3 3 1
author Bob Wilson
author-mail <bob.wilson@example.com>  
author-time 1609632000
	import "fmt"`

	expected := []BlameLine{
		{
			CommitHash:  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
			Author:      "John Doe",
			AuthorEmail: "john.doe@example.com",
			Date:        "1609459200",
			LineNumber:  1,
			Content:     "package main",
		},
		{
			CommitHash:  "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1",
			Author:      "Jane Smith", 
			AuthorEmail: "jane.smith@example.com",
			Date:        "1609545600",
			LineNumber:  2,
			Content:     "",
		},
		{
			CommitHash:  "c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2",
			Author:      "Bob Wilson",
			AuthorEmail: "bob.wilson@example.com",
			Date:        "1609632000",
			LineNumber:  3,
			Content:     "import \"fmt\"",
		},
	}

	result, err := parseGitBlameOutput(sampleOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d lines, got %d", len(expected), len(result))
	}

	for i, line := range result {
		if line.CommitHash != expected[i].CommitHash {
			t.Errorf("line %d: expected commit %s, got %s", i+1, expected[i].CommitHash, line.CommitHash)
		}
		if line.Author != expected[i].Author {
			t.Errorf("line %d: expected author %s, got %s", i+1, expected[i].Author, line.Author)
		}
		if line.AuthorEmail != expected[i].AuthorEmail {
			t.Errorf("line %d: expected email %s, got %s", i+1, expected[i].AuthorEmail, line.AuthorEmail)
		}
		if line.Date != expected[i].Date {
			t.Errorf("line %d: expected date %s, got %s", i+1, expected[i].Date, line.Date)
		}
		if line.LineNumber != expected[i].LineNumber {
			t.Errorf("line %d: expected line number %d, got %d", i+1, expected[i].LineNumber, line.LineNumber)
		}
		if line.Content != expected[i].Content {
			t.Errorf("line %d: expected content %q, got %q", i+1, expected[i].Content, line.Content)
		}
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"a1b2c3d4e5f6", true},
		{"ABCDEF123456", true},
		{"123456789abc", true},
		{"g1b2c3d4e5f6", false},
		{"12345G789abc", false},
		{"", true}, // empty string should be true
		{"xyz", false},
		{"123", true},
	}

	for _, test := range tests {
		result := isHexString(test.input)
		if result != test.expected {
			t.Errorf("isHexString(%q) = %t, expected %t", test.input, result, test.expected)
		}
	}
}

func TestExecuteGitBlameIntegration(t *testing.T) {
	// Skip if not in a git repository
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	repoRoot, err := FindGitRoot(wd)
	if err != nil {
		t.Skipf("skipping integration test: not in git repository: %v", err)
	}

	// Test with this very file
	thisFile := filepath.Join(wd, "git_test.go")
	
	lines, err := ExecuteGitBlame(repoRoot, thisFile, "", false)
	if err != nil {
		t.Fatalf("ExecuteGitBlame failed: %v", err)
	}

	if len(lines) == 0 {
		t.Fatal("expected some blame lines, got none")
	}

	// Verify structure of first line
	firstLine := lines[0]
	if firstLine.CommitHash == "" {
		t.Error("expected commit hash, got empty string")
	}
	if len(firstLine.CommitHash) != 40 {
		t.Errorf("expected 40-char commit hash, got %d chars: %s", len(firstLine.CommitHash), firstLine.CommitHash)
	}
	if firstLine.LineNumber != 1 {
		t.Errorf("expected first line number to be 1, got %d", firstLine.LineNumber)
	}
	if !isHexString(firstLine.CommitHash) {
		t.Errorf("commit hash should be hex string: %s", firstLine.CommitHash)
	}
}

func TestParseRepositoryURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectOwner  string
		expectRepo   string
		expectType   RepositoryType
		expectHost   string
		expectError  bool
	}{
		// GitHub tests
		{
			name:        "GitHub SSH format",
			url:         "git@github.com:owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitHub,
			expectHost:  "github.com",
			expectError: false,
		},
		{
			name:        "GitHub HTTPS format",
			url:         "https://github.com/owner/repo.git",
			expectOwner: "owner", 
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitHub,
			expectHost:  "github.com",
			expectError: false,
		},
		{
			name:        "GitHub HTTP format",
			url:         "http://github.com/owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo", 
			expectType:  RepositoryTypeGitHub,
			expectHost:  "github.com",
			expectError: false,
		},
		{
			name:        "GitHub SSH without .git",
			url:         "git@github.com:owner/repo",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitHub,
			expectHost:  "github.com",
			expectError: false,
		},
		{
			name:        "GitHub HTTPS without .git",
			url:         "https://github.com/owner/repo",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitHub,
			expectHost:  "github.com",
			expectError: false,
		},
		{
			name:        "GitHub with subpath (should take first two)",
			url:         "https://github.com/owner/repo/tree/main",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitHub,
			expectHost:  "github.com",
			expectError: false,
		},
		
		// GitLab.com tests
		{
			name:        "GitLab SSH format",
			url:         "git@gitlab.com:owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.com",
			expectError: false,
		},
		{
			name:        "GitLab HTTPS format", 
			url:         "https://gitlab.com/owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.com",
			expectError: false,
		},
		{
			name:        "GitLab HTTP format",
			url:         "http://gitlab.com/owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.com",
			expectError: false,
		},
		{
			name:        "GitLab SSH without .git",
			url:         "git@gitlab.com:owner/repo",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.com",
			expectError: false,
		},
		{
			name:        "GitLab with subpath",
			url:         "https://gitlab.com/owner/repo/-/tree/main",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.com",
			expectError: false,
		},
		
		// Self-hosted GitLab tests
		{
			name:        "Self-hosted GitLab SSH",
			url:         "git@gitlab.example.com:owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo", 
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.example.com",
			expectError: false,
		},
		{
			name:        "Self-hosted GitLab HTTPS",
			url:         "https://gitlab.example.com/owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.example.com",
			expectError: false,
		},
		{
			name:        "Self-hosted GitLab HTTP",
			url:         "http://gitlab.internal.corp/owner/repo",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectType:  RepositoryTypeGitLab,
			expectHost:  "gitlab.internal.corp",
			expectError: false,
		},
		
		// Error cases
		{
			name:        "invalid path format",
			url:         "git@github.com:justowner",
			expectError: true,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
		{
			name:        "unsupported format",
			url:         "ftp://example.com/owner/repo",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRepositoryURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Owner != tt.expectOwner {
				t.Errorf("expected owner %s, got %s", tt.expectOwner, result.Owner)
			}

			if result.Name != tt.expectRepo {
				t.Errorf("expected repo %s, got %s", tt.expectRepo, result.Name)
			}
			
			if result.Type != tt.expectType {
				t.Errorf("expected type %s, got %s", tt.expectType, result.Type)
			}
			
			if result.Host != tt.expectHost {
				t.Errorf("expected host %s, got %s", tt.expectHost, result.Host)
			}
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectOwner string
		expectRepo  string
		expectError bool
	}{
		{
			name:        "GitHub SSH format",
			url:         "git@github.com:owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectError: false,
		},
		{
			name:        "GitHub HTTPS format",
			url:         "https://github.com/owner/repo.git",
			expectOwner: "owner", 
			expectRepo:  "repo",
			expectError: false,
		},
		{
			name:        "GitLab URL should error",
			url:         "https://gitlab.com/owner/repo.git",
			expectError: true,
		},
		{
			name:        "invalid path format",
			url:         "git@github.com:justowner",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGitHubURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Owner != tt.expectOwner {
				t.Errorf("expected owner %s, got %s", tt.expectOwner, result.Owner)
			}

			if result.Name != tt.expectRepo {
				t.Errorf("expected repo %s, got %s", tt.expectRepo, result.Name)
			}
		})
	}
}

func TestParseRepoPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectOwner string
		expectRepo  string
		expectError bool
	}{
		{
			name:        "basic owner/repo",
			path:        "owner/repo",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectError: false,
		},
		{
			name:        "with .git suffix",
			path:        "owner/repo.git",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectError: false,
		},
		{
			name:        "with additional path segments",
			path:        "owner/repo/tree/main",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectError: false,
		},
		{
			name:        "missing repo name",
			path:        "owner",
			expectError: true,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRepoPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Owner != tt.expectOwner {
				t.Errorf("expected owner %s, got %s", tt.expectOwner, result.Owner)
			}

			if result.Name != tt.expectRepo {
				t.Errorf("expected repo %s, got %s", tt.expectRepo, result.Name)
			}
		})
	}
}