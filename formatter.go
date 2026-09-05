package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// shortHashLength is the number of leading characters of a commit hash shown
// in human-readable output, matching git blame's default abbreviation length.
const shortHashLength = 8

// OutputFormatter handles formatting blame output for display
type OutputFormatter struct {
	ShowEmail bool
	Porcelain bool
	NoColors  bool
}

// BlameLineWithApproval combines blame line with PR approval information
type BlameLineWithApproval struct {
	BlameLine
	PRNumber      int
	Approver      string
	ApproverEmail string
	ApprovalTime  *time.Time
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
		if len(shortHash) > shortHashLength {
			shortHash = shortHash[:shortHashLength]
		}

		// Author name (approver if available, otherwise original author)
		authorName := f.getAuthorName(line)

		// Date (approval time if available, otherwise commit time)
		dateStr := f.getDateString(line)

		// Line number
		lineNumStr := fmt.Sprintf("%*d", maxLineNumWidth, line.LineNumber)

		// Format the line: hash (author date lineNum) content
		fmt.Fprintf(&result, "%s (%-*s %s %s) %s\n",
			shortHash,
			maxAuthorWidth, authorName,
			dateStr,
			lineNumStr,
			line.Content,
		)
	}

	return result.String()
}

// formatPorcelain formats output in porcelain format for machine parsing
func (f *OutputFormatter) formatPorcelain(lines []BlameLineWithApproval) string {
	var result strings.Builder

	for _, line := range lines {
		// Commit hash and line info
		fmt.Fprintf(&result, "%s %d %d 1\n",
			line.CommitHash,
			line.LineNumber,
			line.LineNumber)

		// Author info (use approver if available)
		if line.Approver != "" {
			fmt.Fprintf(&result, "author %s\n", line.Approver)
			if line.ApproverEmail != "" {
				fmt.Fprintf(&result, "author-mail <%s>\n", line.ApproverEmail)
			}
			if line.ApprovalTime != nil {
				fmt.Fprintf(&result, "author-time %d\n", line.ApprovalTime.Unix())
			}
		} else {
			// Fall back to original author
			fmt.Fprintf(&result, "author %s\n", line.Author)
			fmt.Fprintf(&result, "author-mail <%s>\n", line.AuthorEmail)
			if timestamp, err := strconv.ParseInt(line.Date, 10, 64); err == nil {
				fmt.Fprintf(&result, "author-time %d\n", timestamp)
			}
		}

		// Additional PR info
		if line.PRNumber > 0 {
			fmt.Fprintf(&result, "pr-number %d\n", line.PRNumber)
		}

		fmt.Fprintf(&result, "filename %s\n", "") // We don't have filename in context
		fmt.Fprintf(&result, "\t%s\n", line.Content)
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
