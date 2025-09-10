package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		token:   token,
		baseURL: "https://api.github.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an authenticated request to the GitHub API
func (c *GitHubClient) makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return c.httpClient.Do(req)
}

// PullRequest represents basic PR information from GitHub API
type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	MergedAt *time.Time `json:"merged_at"`
}

// Review represents a PR review from GitHub API
type Review struct {
	User struct {
		Login string `json:"login"`
		Email string `json:"email"`
	} `json:"user"`
	State       string     `json:"state"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

// PRApprovalInfo contains information about PR approvals
type PRApprovalInfo struct {
	PR        PullRequest
	Approvers []Review
}

// FindPRByCommit finds the pull request that introduced a specific commit
func (c *GitHubClient) FindPRByCommit(owner, repo, commitHash string) (*PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/pulls", c.baseURL, owner, repo, commitHash)
	
	resp, err := c.makeRequest("GET", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var prs []PullRequest
	if err := json.Unmarshal(body, &prs); err != nil {
		return nil, err
	}

	// Return the first (most relevant) PR, or nil if none found
	if len(prs) == 0 {
		return nil, nil
	}

	return &prs[0], nil
}

// GetPRApprovals gets all approvals for a specific pull request
func (c *GitHubClient) GetPRApprovals(owner, repo string, prNumber int) ([]Review, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews", c.baseURL, owner, repo, prNumber)
	
	resp, err := c.makeRequest("GET", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reviews []Review
	if err := json.Unmarshal(body, &reviews); err != nil {
		return nil, err
	}

	// Filter only approved reviews
	var approvals []Review
	for _, review := range reviews {
		if review.State == "APPROVED" {
			approvals = append(approvals, review)
		}
	}

	return approvals, nil
}

// GetPRApprovalInfo gets complete approval information for a commit
func (c *GitHubClient) GetPRApprovalInfo(owner, repo, commitHash string) (*PRApprovalInfo, error) {
	pr, err := c.FindPRByCommit(owner, repo, commitHash)
	if err != nil {
		return nil, err
	}

	if pr == nil {
		return nil, fmt.Errorf("no pull request found for commit %s", commitHash)
	}

	approvals, err := c.GetPRApprovals(owner, repo, pr.Number)
	if err != nil {
		return nil, err
	}

	return &PRApprovalInfo{
		PR:        *pr,
		Approvers: approvals,
	}, nil
}

// GitHubClientAdapter adapts GitHubClient to implement ReviewClient interface
type GitHubClientAdapter struct {
	client *GitHubClient
}

// NewGitHubClientAdapter creates a new adapter for GitHubClient
func NewGitHubClientAdapter(token string) ReviewClient {
	return &GitHubClientAdapter{
		client: NewGitHubClient(token),
	}
}

// FindPRByCommit implements ReviewClient interface
func (a *GitHubClientAdapter) FindPRByCommit(owner, repo, commitHash string) (*PullRequest, error) {
	return a.client.FindPRByCommit(owner, repo, commitHash)
}

// GetPRApprovals implements ReviewClient interface  
func (a *GitHubClientAdapter) GetPRApprovals(owner, repo string, prNumber int) ([]Review, error) {
	return a.client.GetPRApprovals(owner, repo, prNumber)
}

// GetPRApprovalInfo implements ReviewClient interface
func (a *GitHubClientAdapter) GetPRApprovalInfo(owner, repo, commitHash string) (*PRApprovalInfo, error) {
	return a.client.GetPRApprovalInfo(owner, repo, commitHash)
}