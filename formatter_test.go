package main

import (
	"strings"
	"testing"
	"time"
)

func TestNewOutputFormatter(t *testing.T) {
	formatter := NewOutputFormatter(true, false, true)
	
	if !formatter.ShowEmail {
		t.Error("expected ShowEmail to be true")
	}
	if formatter.Porcelain {
		t.Error("expected Porcelain to be false")
	}
	if !formatter.NoColors {
		t.Error("expected NoColors to be true")
	}
}

func TestFormatHuman(t *testing.T) {
	approvalTime := time.Unix(1609632000, 0)
	
	lines := []BlameLineWithApproval{
		{
			BlameLine: BlameLine{
				CommitHash:  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
				Author:      "John Doe",
				AuthorEmail: "john@example.com",
				Date:        "1609459200",
				LineNumber:  1,
				Content:     "package main",
			},
			PRNumber:      123,
			Approver:      "Jane Smith",
			ApproverEmail: "jane@example.com", 
			ApprovalTime:  &approvalTime,
		},
		{
			BlameLine: BlameLine{
				CommitHash:  "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1",
				Author:      "Bob Wilson", 
				AuthorEmail: "bob@example.com",
				Date:        "1609545600",
				LineNumber:  2,
				Content:     "import \"fmt\"",
			},
			// No PR info - should fall back to original author
		},
	}
	
	formatter := NewOutputFormatter(false, false, true)
	output := formatter.FormatOutput(lines)
	
	// Check that output contains expected elements
	if !strings.Contains(output, "a1b2c3d4") {
		t.Error("expected shortened commit hash in output")
	}
	if !strings.Contains(output, "Jane Smith") {
		t.Error("expected approver name in output")
	}
	if !strings.Contains(output, "Bob Wilson") {
		t.Error("expected fallback to original author")
	}
	if !strings.Contains(output, "package main") {
		t.Error("expected code content in output")
	}
	if !strings.Contains(output, "2021-01-03") {
		t.Error("expected formatted approval time")
	}
	
	// Check line numbers
	if !strings.Contains(output, " 1) ") {
		t.Error("expected line number 1 in output")
	}
	if !strings.Contains(output, " 2) ") {
		t.Error("expected line number 2 in output")
	}
}

func TestFormatHumanWithEmail(t *testing.T) {
	lines := []BlameLineWithApproval{
		{
			BlameLine: BlameLine{
				CommitHash:  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
				Author:      "John Doe",
				AuthorEmail: "john@example.com",
				Date:        "1609459200",
				LineNumber:  1,
				Content:     "package main",
			},
			Approver:      "Jane Smith",
			ApproverEmail: "jane@example.com",
		},
	}
	
	formatter := NewOutputFormatter(true, false, true) // ShowEmail = true
	output := formatter.FormatOutput(lines)
	
	if !strings.Contains(output, "jane@example.com") {
		t.Error("expected approver email in output when ShowEmail=true")
	}
	if strings.Contains(output, "Jane Smith") {
		t.Error("expected email instead of name when ShowEmail=true")
	}
}

func TestFormatPorcelain(t *testing.T) {
	approvalTime := time.Unix(1609632000, 0)
	
	lines := []BlameLineWithApproval{
		{
			BlameLine: BlameLine{
				CommitHash:  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
				Author:      "John Doe",
				AuthorEmail: "john@example.com", 
				Date:        "1609459200",
				LineNumber:  1,
				Content:     "package main",
			},
			PRNumber:      123,
			Approver:      "Jane Smith",
			ApproverEmail: "jane@example.com",
			ApprovalTime:  &approvalTime,
		},
	}
	
	formatter := NewOutputFormatter(false, true, true) // Porcelain = true
	output := formatter.FormatOutput(lines)
	
	expectedLines := []string{
		"a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0 1 1 1",
		"author Jane Smith",
		"author-mail <jane@example.com>",
		"author-time 1609632000", 
		"pr-number 123",
		"\tpackage main",
	}
	
	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("expected %q in porcelain output, got:\n%s", expected, output)
		}
	}
}

func TestFormatPorcelainFallback(t *testing.T) {
	lines := []BlameLineWithApproval{
		{
			BlameLine: BlameLine{
				CommitHash:  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
				Author:      "John Doe",
				AuthorEmail: "john@example.com",
				Date:        "1609459200", 
				LineNumber:  1,
				Content:     "package main",
			},
			// No approver info - should use original author
		},
	}
	
	formatter := NewOutputFormatter(false, true, true)
	output := formatter.FormatOutput(lines)
	
	expectedLines := []string{
		"author John Doe",
		"author-mail <john@example.com>",
		"author-time 1609459200",
	}
	
	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("expected %q in porcelain fallback output, got:\n%s", expected, output)
		}
	}
	
	// Should not contain PR info
	if strings.Contains(output, "pr-number") {
		t.Error("should not contain pr-number when no PR info available")
	}
}

func TestGetAuthorName(t *testing.T) {
	formatter := NewOutputFormatter(false, false, false)
	
	// Test with approver info
	lineWithApprover := BlameLineWithApproval{
		BlameLine: BlameLine{
			Author:      "John Doe",
			AuthorEmail: "john@example.com",
		},
		Approver:      "Jane Smith", 
		ApproverEmail: "jane@example.com",
	}
	
	name := formatter.getAuthorName(lineWithApprover)
	if name != "Jane Smith" {
		t.Errorf("expected 'Jane Smith', got '%s'", name)
	}
	
	// Test fallback to original author
	lineWithoutApprover := BlameLineWithApproval{
		BlameLine: BlameLine{
			Author:      "John Doe",
			AuthorEmail: "john@example.com",
		},
	}
	
	name = formatter.getAuthorName(lineWithoutApprover)
	if name != "John Doe" {
		t.Errorf("expected 'John Doe', got '%s'", name)
	}
	
	// Test with ShowEmail
	formatter.ShowEmail = true
	name = formatter.getAuthorName(lineWithApprover)
	if name != "jane@example.com" {
		t.Errorf("expected 'jane@example.com', got '%s'", name)
	}
}

func TestGetDateString(t *testing.T) {
	formatter := NewOutputFormatter(false, false, false)
	
	approvalTime := time.Unix(1609632000, 0)
	
	// Test with approval time
	lineWithApproval := BlameLineWithApproval{
		BlameLine: BlameLine{
			Date: "1609459200",
		},
		ApprovalTime: &approvalTime,
	}
	
	dateStr := formatter.getDateString(lineWithApproval)
	if !strings.Contains(dateStr, "2021-01-03") {
		t.Errorf("expected formatted approval time, got '%s'", dateStr)
	}
	
	// Test fallback to commit date
	lineWithoutApproval := BlameLineWithApproval{
		BlameLine: BlameLine{
			Date: "1609459200",
		},
	}
	
	dateStr = formatter.getDateString(lineWithoutApproval)
	if !strings.Contains(dateStr, "2021-01-01") {
		t.Errorf("expected formatted commit time, got '%s'", dateStr)
	}
}

func TestFormatOutputEmpty(t *testing.T) {
	formatter := NewOutputFormatter(false, false, false)
	output := formatter.FormatOutput([]BlameLineWithApproval{})
	
	if output != "" {
		t.Errorf("expected empty output for empty input, got '%s'", output)
	}
}