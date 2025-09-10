package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewGitHubClient(t *testing.T) {
	token := "test-token"
	client := NewGitHubClient(token)

	if client.token != token {
		t.Errorf("expected token %s, got %s", token, client.token)
	}

	if client.baseURL != "https://api.github.com" {
		t.Errorf("expected baseURL https://api.github.com, got %s", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestMakeRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got %s", auth)
		}

		accept := r.Header.Get("Accept")
		if accept != "application/vnd.github+json" {
			t.Errorf("expected Accept header 'application/vnd.github+json', got %s", accept)
		}

		apiVersion := r.Header.Get("X-GitHub-Api-Version")
		if apiVersion != "2022-11-28" {
			t.Errorf("expected X-GitHub-Api-Version header '2022-11-28', got %s", apiVersion)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := NewGitHubClient("test-token")
	resp, err := client.makeRequest("GET", server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestFindPRByCommit(t *testing.T) {
	// Mock PR data
	mockPRs := []PullRequest{
		{
			Number: 123,
			Title:  "Test PR",
			State:  "merged",
		},
	}
	mockPRs[0].User.Login = "testuser"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/repos/owner/repo/commits/abc123/pulls"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockPRs)
	}))
	defer server.Close()

	client := NewGitHubClient("test-token")
	client.baseURL = server.URL

	pr, err := client.FindPRByCommit("owner", "repo", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pr == nil {
		t.Fatal("expected PR, got nil")
	}

	if pr.Number != 123 {
		t.Errorf("expected PR number 123, got %d", pr.Number)
	}

	if pr.Title != "Test PR" {
		t.Errorf("expected PR title 'Test PR', got %s", pr.Title)
	}
}

func TestFindPRByCommitNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]PullRequest{}) // Empty array
	}))
	defer server.Close()

	client := NewGitHubClient("test-token")
	client.baseURL = server.URL

	pr, err := client.FindPRByCommit("owner", "repo", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pr != nil {
		t.Errorf("expected nil PR, got %+v", pr)
	}
}

func TestGetPRApprovals(t *testing.T) {
	mockReviews := []Review{
		{State: "APPROVED"},
		{State: "APPROVED"},
		{State: "COMMENTED"},
	}
	mockReviews[0].User.Login = "approver1"
	mockReviews[0].User.Email = "approver1@example.com"
	mockReviews[1].User.Login = "approver2"
	mockReviews[1].User.Email = "approver2@example.com"
	mockReviews[2].User.Login = "commenter"
	mockReviews[2].User.Email = "commenter@example.com"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/repos/owner/repo/pulls/123/reviews"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockReviews)
	}))
	defer server.Close()

	client := NewGitHubClient("test-token")
	client.baseURL = server.URL

	approvals, err := client.GetPRApprovals("owner", "repo", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only return approved reviews (2 out of 3)
	if len(approvals) != 2 {
		t.Errorf("expected 2 approvals, got %d", len(approvals))
	}

	for _, approval := range approvals {
		if approval.State != "APPROVED" {
			t.Errorf("expected approval state APPROVED, got %s", approval.State)
		}
	}
}

func TestGetPRApprovalInfo(t *testing.T) {
	mockPRs := []PullRequest{
		{
			Number: 123,
			Title:  "Test PR",
			State:  "merged",
		},
	}

	mockReviews := []Review{
		{State: "APPROVED"},
	}
	mockReviews[0].User.Login = "approver1"
	mockReviews[0].User.Email = "approver1@example.com"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/repos/owner/repo/commits/abc123/pulls" {
			json.NewEncoder(w).Encode(mockPRs)
		} else if r.URL.Path == "/repos/owner/repo/pulls/123/reviews" {
			json.NewEncoder(w).Encode(mockReviews)
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewGitHubClient("test-token")
	client.baseURL = server.URL

	info, err := client.GetPRApprovalInfo("owner", "repo", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info == nil {
		t.Fatal("expected approval info, got nil")
	}

	if info.PR.Number != 123 {
		t.Errorf("expected PR number 123, got %d", info.PR.Number)
	}

	if len(info.Approvers) != 1 {
		t.Errorf("expected 1 approver, got %d", len(info.Approvers))
	}

	if info.Approvers[0].User.Login != "approver1" {
		t.Errorf("expected approver 'approver1', got %s", info.Approvers[0].User.Login)
	}
}