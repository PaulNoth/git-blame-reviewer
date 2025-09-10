package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// OutputFormatter handles formatting blame output for display
type OutputFormatter struct {
	ShowEmail  bool
	Porcelain  bool
	NoColors   bool
}

// BlameLineWithApproval combines blame line with PR approval information
type BlameLineWithApproval struct {
	BlameLine
	PRNumber    int
	Approver    string
	ApproverEmail string
	ApprovalTime *time.Time
}

// FormatOutput formats the blame lines with approval information for display
func (f *OutputFormatter) FormatOutput(lines []BlameLineWithApproval) string {
	if f.Porcelain {
		return f.formatPorcelain(lines)
	}
	return f.formatHuman(lines)
}

// formatHuman formats output in human-readable format similar to git blame
func (f *OutputFormatter) formatHuman(lines []BlameLineWithApproval) string {
	if len(lines) == 0 {
		return ""
	}

	var result strings.Builder
	
	// Calculate maximum widths for alignment
	maxAuthorWidth := 0
	maxLineNumWidth := len(strconv.Itoa(len(lines)))
	
	for _, line := range lines {
		authorName := f.getAuthorName(line)
		if len(authorName) > maxAuthorWidth {
			maxAuthorWidth = len(authorName)
		}
	}
	
	// Format each line
	for _, line := range lines {
		// Commit hash (shortened to 8 chars)
		shortHash := line.CommitHash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}
		
		// Author name (approver if available, otherwise original author)
		authorName := f.getAuthorName(line)
		
		// Date (approval time if available, otherwise commit time)
		dateStr := f.getDateString(line)
		
		// Line number
		lineNumStr := fmt.Sprintf("%*d", maxLineNumWidth, line.LineNumber)
		
		// Format the line: hash (author date lineNum) content
		result.WriteString(fmt.Sprintf("%s (%-*s %s %s) %s\n",
			shortHash,
			maxAuthorWidth, authorName,
			dateStr,
			lineNumStr,
			line.Content,
		))
	}
	
	return result.String()
}

// formatPorcelain formats output in porcelain format for machine parsing
func (f *OutputFormatter) formatPorcelain(lines []BlameLineWithApproval) string {
	var result strings.Builder
	
	for _, line := range lines {
		// Commit hash and line info
		result.WriteString(fmt.Sprintf("%s %d %d 1\n", 
			line.CommitHash, 
			line.LineNumber, 
			line.LineNumber))
		
		// Author info (use approver if available)
		if line.Approver != "" {
			result.WriteString(fmt.Sprintf("author %s\n", line.Approver))
			if line.ApproverEmail != "" {
				result.WriteString(fmt.Sprintf("author-mail <%s>\n", line.ApproverEmail))
			}
			if line.ApprovalTime != nil {
				result.WriteString(fmt.Sprintf("author-time %d\n", line.ApprovalTime.Unix()))
			}
		} else {
			// Fall back to original author
			result.WriteString(fmt.Sprintf("author %s\n", line.Author))
			result.WriteString(fmt.Sprintf("author-mail <%s>\n", line.AuthorEmail))
			if timestamp, err := strconv.ParseInt(line.Date, 10, 64); err == nil {
				result.WriteString(fmt.Sprintf("author-time %d\n", timestamp))
			}
		}
		
		// Additional PR info
		if line.PRNumber > 0 {
			result.WriteString(fmt.Sprintf("pr-number %d\n", line.PRNumber))
		}
		
		result.WriteString(fmt.Sprintf("filename %s\n", "")) // We don't have filename in context
		result.WriteString(fmt.Sprintf("\t%s\n", line.Content))
	}
	
	return result.String()
}

// getAuthorName returns the appropriate author name (approver preferred)
func (f *OutputFormatter) getAuthorName(line BlameLineWithApproval) string {
	if line.Approver != "" {
		if f.ShowEmail && line.ApproverEmail != "" {
			return line.ApproverEmail
		}
		return line.Approver
	}
	
	if f.ShowEmail && line.AuthorEmail != "" {
		return line.AuthorEmail
	}
	return line.Author
}

// getDateString returns formatted date string (approval time preferred)
func (f *OutputFormatter) getDateString(line BlameLineWithApproval) string {
	if line.ApprovalTime != nil {
		return line.ApprovalTime.Format("2006-01-02 15:04:05")
	}
	
	// Try to parse original commit date
	if timestamp, err := strconv.ParseInt(line.Date, 10, 64); err == nil {
		return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
	}
	
	return line.Date
}

// NewOutputFormatter creates a new formatter with the given options
func NewOutputFormatter(showEmail, porcelain, noColors bool) *OutputFormatter {
	return &OutputFormatter{
		ShowEmail: showEmail,
		Porcelain: porcelain,
		NoColors:  noColors,
	}
}