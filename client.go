package main

import "time"

// ReviewClient defines the interface for both GitHub and GitLab API clients
type ReviewClient interface {
	// FindPRByCommit finds the pull/merge request that introduced a specific commit
	FindPRByCommit(owner, repo, commitHash string) (*PullRequest, error)
	
	// GetPRApprovals gets all approvals for a specific pull/merge request
	GetPRApprovals(owner, repo string, prNumber int) ([]Review, error)
	
	// GetPRApprovalInfo gets complete approval information for a commit
	GetPRApprovalInfo(owner, repo, commitHash string) (*PRApprovalInfo, error)
}

// UnifiedPullRequest represents a PR/MR from either GitHub or GitLab
type UnifiedPullRequest struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	State    string `json:"state"`
	User     User   `json:"user"`
	MergedAt *time.Time `json:"merged_at"`
	WebURL   string `json:"web_url"` // GitLab uses web_url
}

// User represents a user from either GitHub or GitLab
type User struct {
	Login string `json:"login"` // GitHub uses "login"
	Name  string `json:"name"`  // GitLab uses "name" 
	Email string `json:"email"`
}

// UnifiedReview represents a review/approval from either GitHub or GitLab
type UnifiedReview struct {
	User        User       `json:"user"`
	State       string     `json:"state"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

// UnifiedPRApprovalInfo contains unified information about PR/MR approvals
type UnifiedPRApprovalInfo struct {
	PR        UnifiedPullRequest
	Approvers []UnifiedReview
}

// ClientFactory creates the appropriate client based on repository type
type ClientFactory struct{}

// NewClientFactory creates a new client factory
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// CreateClient creates the appropriate client based on repository type and token availability
func (cf *ClientFactory) CreateClient(repoInfo *RepoInfo, githubToken, gitlabToken string) (ReviewClient, error) {
	switch repoInfo.Type {
	case RepositoryTypeGitHub:
		if githubToken == "" {
			return nil, ErrMissingGitHubToken
		}
		return NewGitHubClientAdapter(githubToken), nil
	case RepositoryTypeGitLab:
		if gitlabToken == "" {
			return nil, ErrMissingGitLabToken
		}
		return NewGitLabClient(gitlabToken, repoInfo.Host), nil
	default:
		return nil, ErrUnsupportedRepositoryType
	}
}

// Custom errors
var (
	ErrMissingGitHubToken        = &ClientError{Message: "GITHUB_TOKEN environment variable is required for GitHub repositories"}
	ErrMissingGitLabToken        = &ClientError{Message: "GITLAB_TOKEN environment variable is required for GitLab repositories"}
	ErrUnsupportedRepositoryType = &ClientError{Message: "unsupported repository type"}
)

// ClientError represents a client-related error
type ClientError struct {
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}