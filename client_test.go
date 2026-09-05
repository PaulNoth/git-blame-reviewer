package main

import (
	"testing"
)

func TestClientFactory(t *testing.T) {
	factory := NewClientFactory()

	tests := []struct {
		name         string
		repoInfo     *RepoInfo
		githubToken  string
		gitlabToken  string
		expectError  bool
		expectClient bool
	}{
		{
			name: "GitHub repository with token",
			repoInfo: &RepoInfo{
				Owner: "owner",
				Name:  "repo",
				Type:  RepositoryTypeGitHub,
				Host:  "github.com",
			},
			githubToken:  "github-token",
			expectError:  false,
			expectClient: true,
		},
		{
			name: "GitHub repository without token",
			repoInfo: &RepoInfo{
				Owner: "owner",
				Name:  "repo",
				Type:  RepositoryTypeGitHub,
				Host:  "github.com",
			},
			expectError:  true,
			expectClient: false,
		},
		{
			name: "GitLab repository with token",
			repoInfo: &RepoInfo{
				Owner: "owner",
				Name:  "repo",
				Type:  RepositoryTypeGitLab,
				Host:  "gitlab.com",
			},
			gitlabToken:  "gitlab-token",
			expectError:  false,
			expectClient: true,
		},
		{
			name: "GitLab repository without token",
			repoInfo: &RepoInfo{
				Owner: "owner",
				Name:  "repo",
				Type:  RepositoryTypeGitLab,
				Host:  "gitlab.com",
			},
			expectError:  true,
			expectClient: false,
		},
		{
			name: "Self-hosted GitLab with token",
			repoInfo: &RepoInfo{
				Owner: "owner",
				Name:  "repo",
				Type:  RepositoryTypeGitLab,
				Host:  "gitlab.example.com",
			},
			gitlabToken:  "gitlab-token",
			expectError:  false,
			expectClient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := factory.CreateClient(tt.repoInfo, tt.githubToken, tt.gitlabToken)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if client != nil {
					t.Errorf("expected nil client but got %T", client)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectClient {
				if client == nil {
					t.Error("expected client but got nil")
					return
				}

				// client is already statically typed as ReviewClient by
				// CreateClient's signature; nothing further to assert here
				// beyond the non-nil check above.
			}
		})
	}
}

func TestGitHubClientAdapter(t *testing.T) {
	// NewGitHubClientAdapter's return type already guarantees ReviewClient
	// conformance at compile time; just verify a non-nil client is created.
	adapter := NewGitHubClientAdapter("test-token")

	if adapter == nil {
		t.Error("expected adapter to be created")
	}
}

func TestGitLabClient(t *testing.T) {
	// NewGitLabClient's return type already guarantees ReviewClient
	// conformance at compile time; just verify a non-nil client is created.
	client := NewGitLabClient("test-token", "gitlab.com")

	if client == nil {
		t.Error("expected client to be created")
	}
}

func TestClientError(t *testing.T) {
	err := &ClientError{Message: "test error"}

	expected := "test error"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestRepositoryTypeString(t *testing.T) {
	tests := []struct {
		repoType RepositoryType
		expected string
	}{
		{RepositoryTypeGitHub, "GitHub"},
		{RepositoryTypeGitLab, "GitLab"},
		{RepositoryType(999), "Unknown"},
	}

	for _, tt := range tests {
		result := tt.repoType.String()
		if result != tt.expected {
			t.Errorf("expected %s.String() = %q, got %q", tt.repoType, tt.expected, result)
		}
	}
}

// TestReviewClientInterface ensures both clients implement the interface correctly
func TestReviewClientInterface(t *testing.T) {
	clients := []struct {
		name   string
		client ReviewClient
	}{
		{
			name:   "GitHubClientAdapter",
			client: NewGitHubClientAdapter("test-token"),
		},
		{
			name:   "GitLabClient",
			client: NewGitLabClient("test-token", "gitlab.com"),
		},
	}

	for _, tc := range clients {
		t.Run(tc.name, func(t *testing.T) {
			// Test that all interface methods exist and can be called
			// (We can't test actual functionality without real API calls)

			if tc.client == nil {
				t.Fatal("client is nil")
			}

			// Test method signatures exist (will compile if interface is correct)
			var _ = tc.client.FindPRByCommit
			var _ = tc.client.GetPRApprovals
			var _ = tc.client.GetPRApprovalInfo
		})
	}
}
