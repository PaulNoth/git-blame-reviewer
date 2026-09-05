package main

import (
	"errors"
	"testing"
	"time"
)

// fakeReviewClient is a deterministic, in-memory ReviewClient used to test
// the "real success" path of buildBlameLinesWithApprovals without any
// network calls or API tokens.
type fakeReviewClient struct {
	// approvals maps a commit hash to the approval info that should be
	// returned for it. Commits not present in this map cause an error to be
	// returned, simulating "no PR found for this commit".
	approvals map[string]*PRApprovalInfo
}

func (f *fakeReviewClient) FindPRByCommit(_, _, commitHash string) (*PullRequest, error) {
	if info, ok := f.approvals[commitHash]; ok {
		return &info.PR, nil
	}
	return nil, errors.New("no PR found for commit")
}

func (f *fakeReviewClient) GetPRApprovals(_, _ string, _ int) ([]Review, error) {
	return nil, errors.New("not implemented in fake")
}

func (f *fakeReviewClient) GetPRApprovalInfo(_, _, commitHash string) (*PRApprovalInfo, error) {
	if info, ok := f.approvals[commitHash]; ok {
		return info, nil
	}
	return nil, errors.New("no PR found for commit")
}

func TestBuildBlameLinesWithApprovals_Success(t *testing.T) {
	approvedAt := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)

	var review Review
	review.User.Login = "alice"
	review.User.Email = "alice@example.com"
	review.State = "APPROVED"
	review.SubmittedAt = &approvedAt

	approvalInfo := &PRApprovalInfo{
		PR:        PullRequest{Number: 42, State: "closed"},
		Approvers: []Review{review},
	}

	client := &fakeReviewClient{
		approvals: map[string]*PRApprovalInfo{
			"abc123": approvalInfo,
		},
	}

	blameLines := []BlameLine{
		{
			CommitHash:  "abc123",
			Author:      "original-author",
			AuthorEmail: "original@example.com",
			Date:        "1700000000",
			LineNumber:  1,
			Content:     "package main",
		},
	}

	repoInfo := &RepoInfo{Owner: "PaulNoth", Name: "git-blame-reviewer", Type: RepositoryTypeGitHub}

	result := buildBlameLinesWithApprovals(blameLines, client, repoInfo)

	if len(result) != 1 {
		t.Fatalf("expected 1 result line, got %d", len(result))
	}

	line := result[0]
	if line.PRNumber != 42 {
		t.Errorf("expected PRNumber 42, got %d", line.PRNumber)
	}
	if line.Approver != "alice" {
		t.Errorf("expected Approver %q, got %q", "alice", line.Approver)
	}
	if line.ApproverEmail != "alice@example.com" {
		t.Errorf("expected ApproverEmail %q, got %q", "alice@example.com", line.ApproverEmail)
	}
	if line.ApprovalTime == nil || !line.ApprovalTime.Equal(approvedAt) {
		t.Errorf("expected ApprovalTime %v, got %v", approvedAt, line.ApprovalTime)
	}
	// Original blame info should still be preserved alongside approval info.
	if line.Author != "original-author" {
		t.Errorf("expected original Author to be preserved, got %q", line.Author)
	}
}

func TestBuildBlameLinesWithApprovals_FallbackOnLookupFailure(t *testing.T) {
	// No approvals configured, so the fake client returns an error for every
	// commit, simulating an invalid token or a commit with no associated PR.
	client := &fakeReviewClient{approvals: map[string]*PRApprovalInfo{}}

	blameLines := []BlameLine{
		{
			CommitHash:  "def456",
			Author:      "original-author",
			AuthorEmail: "original@example.com",
			Date:        "1700000000",
			LineNumber:  1,
			Content:     "package main",
		},
	}

	repoInfo := &RepoInfo{Owner: "PaulNoth", Name: "git-blame-reviewer", Type: RepositoryTypeGitHub}

	result := buildBlameLinesWithApprovals(blameLines, client, repoInfo)

	if len(result) != 1 {
		t.Fatalf("expected 1 result line, got %d", len(result))
	}

	line := result[0]
	if line.PRNumber != 0 {
		t.Errorf("expected no PRNumber on lookup failure, got %d", line.PRNumber)
	}
	if line.Approver != "" {
		t.Errorf("expected no Approver on lookup failure, got %q", line.Approver)
	}
	// Original blame info must remain untouched so the formatter can fall
	// back to it.
	if line.Author != "original-author" {
		t.Errorf("expected original Author to be preserved, got %q", line.Author)
	}
	if line.Date != "1700000000" {
		t.Errorf("expected original Date to be preserved, got %q", line.Date)
	}
}

func TestBuildBlameLinesWithApprovals_CachesPerCommit(t *testing.T) {
	callCount := 0
	client := &countingReviewClient{
		fakeReviewClient: fakeReviewClient{
			approvals: map[string]*PRApprovalInfo{
				"shared-commit": {PR: PullRequest{Number: 7}},
			},
		},
		calls: &callCount,
	}

	blameLines := []BlameLine{
		{CommitHash: "shared-commit", LineNumber: 1, Content: "line one"},
		{CommitHash: "shared-commit", LineNumber: 2, Content: "line two"},
	}

	repoInfo := &RepoInfo{Owner: "PaulNoth", Name: "git-blame-reviewer", Type: RepositoryTypeGitHub}

	result := buildBlameLinesWithApprovals(blameLines, client, repoInfo)

	if len(result) != 2 {
		t.Fatalf("expected 2 result lines, got %d", len(result))
	}
	if callCount != 1 {
		t.Errorf("expected exactly 1 API call due to per-commit caching, got %d", callCount)
	}
	for i, line := range result {
		if line.PRNumber != 7 {
			t.Errorf("line %d: expected PRNumber 7, got %d", i, line.PRNumber)
		}
	}
}

// countingReviewClient wraps fakeReviewClient to count GetPRApprovalInfo
// calls, used to verify the commit-level cache avoids duplicate lookups.
type countingReviewClient struct {
	fakeReviewClient
	calls *int
}

func (c *countingReviewClient) GetPRApprovalInfo(owner, repo, commitHash string) (*PRApprovalInfo, error) {
	*c.calls++
	return c.fakeReviewClient.GetPRApprovalInfo(owner, repo, commitHash)
}
