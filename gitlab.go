package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GitLabClient handles GitLab API interactions
type GitLabClient struct {
	token      string
	httpClient *http.Client
	baseURL    string
	host       string
}

// NewGitLabClient creates a new GitLab API client
func NewGitLabClient(token, host string) ReviewClient {
	baseURL := fmt.Sprintf("https://%s/api/v4", host)
	if host == "gitlab.com" {
		baseURL = "https://gitlab.com/api/v4"
	}
	
	return &GitLabClient{
		token:   token,
		baseURL: baseURL,
		host:    host,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an authenticated request to the GitLab API
func (c *GitLabClient) makeRequest(method, apiURL string) (*http.Response, error) {
	req, err := http.NewRequest(method, apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication header
	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// GitLabMergeRequest represents basic MR information from GitLab API
type GitLabMergeRequest struct {
	IID       int    `json:"iid"`
	Title     string `json:"title"`
	State     string `json:"state"`
	WebURL    string `json:"web_url"`
	Author    GitLabUser `json:"author"`
	MergedAt  *time.Time `json:"merged_at"`
}

// GitLabUser represents a GitLab user
type GitLabUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GitLabApproval represents a GitLab MR approval
type GitLabApproval struct {
	User GitLabUser `json:"user"`
	CreatedAt *time.Time `json:"created_at"`
}

// FindPRByCommit finds the merge request that introduced a specific commit
func (c *GitLabClient) FindPRByCommit(owner, repo, commitHash string) (*PullRequest, error) {
	// Encode the project path
	projectPath := url.PathEscape(fmt.Sprintf("%s/%s", owner, repo))
	apiURL := fmt.Sprintf("%s/projects/%s/repository/commits/%s/merge_requests", c.baseURL, projectPath, commitHash)
	
	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mrs []GitLabMergeRequest
	if err := json.Unmarshal(body, &mrs); err != nil {
		return nil, err
	}

	// Return the first (most relevant) MR, or nil if none found
	if len(mrs) == 0 {
		return nil, nil
	}

	// Convert GitLab MR to GitHub PR format for compatibility
	mr := mrs[0]
	return &PullRequest{
		Number: mr.IID,
		Title:  mr.Title,
		State:  mr.State,
		User: struct {
			Login string `json:"login"`
		}{Login: mr.Author.Username},
		MergedAt: mr.MergedAt,
	}, nil
}

// GetPRApprovals gets all approvals for a specific merge request
func (c *GitLabClient) GetPRApprovals(owner, repo string, prNumber int) ([]Review, error) {
	// Encode the project path
	projectPath := url.PathEscape(fmt.Sprintf("%s/%s", owner, repo))
	apiURL := fmt.Sprintf("%s/projects/%s/merge_requests/%d/approvals", c.baseURL, projectPath, prNumber)
	
	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// GitLab approval response structure
	var approvalResp struct {
		ApprovedBy []GitLabApproval `json:"approved_by"`
	}
	
	if err := json.Unmarshal(body, &approvalResp); err != nil {
		return nil, err
	}

	// Convert GitLab approvals to GitHub review format
	var reviews []Review
	for _, approval := range approvalResp.ApprovedBy {
		review := Review{
			State:       "APPROVED",
			SubmittedAt: approval.CreatedAt,
		}
		review.User.Login = approval.User.Username
		review.User.Email = approval.User.Email
		reviews = append(reviews, review)
	}

	return reviews, nil
}

// GetPRApprovalInfo gets complete approval information for a commit
func (c *GitLabClient) GetPRApprovalInfo(owner, repo, commitHash string) (*PRApprovalInfo, error) {
	pr, err := c.FindPRByCommit(owner, repo, commitHash)
	if err != nil {
		return nil, err
	}

	if pr == nil {
		return nil, fmt.Errorf("no merge request found for commit %s", commitHash)
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