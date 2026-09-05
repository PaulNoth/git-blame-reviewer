package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNotGitRepo = errors.New("not a git repository")

// splitIntoTwoParts is used with strings.SplitN/expected-length checks when
// splitting a "host@path"/"owner/repo"-shaped string into exactly two parts.
const splitIntoTwoParts = 2

// BlameLine represents a single line from git blame output
type BlameLine struct {
	CommitHash  string
	Author      string
	AuthorEmail string
	Date        string
	LineNumber  int
	Content     string
}

// FindGitRoot finds the root directory of a git repository by walking up
// the directory tree looking for a .git directory
func FindGitRoot(startPath string) (string, error) {
	// Convert to absolute path to handle relative paths consistently
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	// Start from the directory containing the file (if startPath is a file)
	currentPath := absPath
	if info, err := os.Stat(currentPath); err == nil && !info.IsDir() {
		currentPath = filepath.Dir(currentPath)
	}

	// Walk up the directory tree
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			// Check if it's a directory (.git folder) or file (.git file for worktrees)
			if info.IsDir() || info.Mode().IsRegular() {
				return currentPath, nil
			}
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)

		// If we reached the root directory, stop
		if parentPath == currentPath {
			break
		}

		currentPath = parentPath
	}

	return "", ErrNotGitRepo
}

// ExecuteGitBlame runs git blame on the specified file and returns the parsed output
func ExecuteGitBlame(ctx context.Context, repoRoot, filePath, lineRange string, porcelain bool) ([]BlameLine, error) {
	// Build git blame command
	args := []string{"blame"}

	// Add line range if specified
	if lineRange != "" {
		args = append(args, "-L", lineRange)
	}

	// Add porcelain format for easier parsing
	if porcelain {
		args = append(args, "--porcelain")
	} else {
		// Use line porcelain for consistent parsing
		args = append(args, "--line-porcelain")
	}

	// Convert filePath to absolute path first to handle relative paths correctly
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Add the file path (relative to repo root)
	relPath, err := filepath.Rel(repoRoot, absFilePath)
	if err != nil {
		return nil, err
	}
	args = append(args, relPath)

	// Execute git blame
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGitBlameOutput(string(output))
}

// parseGitBlameOutput parses the porcelain output from git blame
func parseGitBlameOutput(output string) ([]BlameLine, error) {
	var lines []BlameLine
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentLine BlameLine
	var lineNumber int

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if this is a commit hash line (starts with hash)
		if len(line) >= 40 && isHexString(line[:40]) {
			// If we have a previous line, save it
			if currentLine.CommitHash != "" {
				lines = append(lines, currentLine)
			}

			// Start new blame line
			parts := strings.Fields(line)
			currentLine = BlameLine{
				CommitHash: parts[0],
				LineNumber: lineNumber + 1,
			}
			lineNumber++
			continue
		}

		// Parse metadata fields
		if strings.HasPrefix(line, "author ") {
			currentLine.Author = line[7:]
		} else if strings.HasPrefix(line, "author-mail ") {
			email := strings.TrimSpace(line[12:])
			// Remove < and > from email
			if len(email) > 2 && email[0] == '<' && email[len(email)-1] == '>' {
				email = email[1 : len(email)-1]
			}
			currentLine.AuthorEmail = email
		} else if strings.HasPrefix(line, "author-time ") {
			currentLine.Date = line[12:]
		} else if strings.HasPrefix(line, "\t") {
			// This is the actual code line (starts with tab)
			currentLine.Content = line[1:] // Remove the leading tab
		}
	}

	// Don't forget the last line
	if currentLine.CommitHash != "" {
		lines = append(lines, currentLine)
	}

	return lines, scanner.Err()
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

// RepositoryType represents the type of git hosting service
type RepositoryType int

const (
	RepositoryTypeGitHub RepositoryType = iota
	RepositoryTypeGitLab
)

func (rt RepositoryType) String() string {
	switch rt {
	case RepositoryTypeGitHub:
		return "GitHub"
	case RepositoryTypeGitLab:
		return "GitLab"
	default:
		return "Unknown"
	}
}

// RepoInfo contains repository owner, name, and type information
type RepoInfo struct {
	Owner string
	Name  string
	Type  RepositoryType
	Host  string // For self-hosted GitLab instances
}

// ExtractRepoInfo extracts owner and repository name from git remote
func ExtractRepoInfo(ctx context.Context, repoRoot string) (*RepoInfo, error) {
	// Get remote origin URL
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	remoteURL := strings.TrimSpace(string(output))

	return parseRepositoryURL(remoteURL)
}

// knownHostPrefix describes one recognized "prefix -> (host, type)" mapping
// for the public GitHub.com/GitLab.com URL formats (SSH, HTTPS, and HTTP).
type knownHostPrefix struct {
	prefix string
	host   string
	rtype  RepositoryType
}

// knownHostPrefixes lists every recognized public GitHub/GitLab URL prefix.
// Order doesn't matter: prefixes are mutually exclusive.
var knownHostPrefixes = []knownHostPrefix{
	{prefix: "git@github.com:", host: githubComHost, rtype: RepositoryTypeGitHub},
	{prefix: "https://github.com/", host: githubComHost, rtype: RepositoryTypeGitHub},
	{prefix: "http://github.com/", host: githubComHost, rtype: RepositoryTypeGitHub},
	{prefix: "git@gitlab.com:", host: gitlabComHost, rtype: RepositoryTypeGitLab},
	{prefix: "https://gitlab.com/", host: gitlabComHost, rtype: RepositoryTypeGitLab},
	{prefix: "http://gitlab.com/", host: gitlabComHost, rtype: RepositoryTypeGitLab},
}

// parseRepositoryURL extracts owner, repo name, and type from GitHub/GitLab URLs
func parseRepositoryURL(url string) (*RepoInfo, error) {
	url = strings.TrimSpace(url)

	if repoInfo, matched, err := parseKnownHostURL(url); matched {
		return repoInfo, err
	}

	if repoInfo, matched, err := parseSelfHostedSSHURL(url); matched {
		return repoInfo, err
	}

	if repoInfo, matched, err := parseSelfHostedHTTPURL(url); matched {
		return repoInfo, err
	}

	return nil, fmt.Errorf("unsupported repository URL format: %s", url)
}

// parseKnownHostURL handles the public GitHub.com/GitLab.com SSH/HTTPS/HTTP
// URL formats, e.g. "git@github.com:owner/repo.git" or
// "https://gitlab.com/owner/repo.git". The bool return indicates whether url
// matched one of the known prefixes at all.
func parseKnownHostURL(url string) (repoInfo *RepoInfo, matched bool, err error) {
	for _, known := range knownHostPrefixes {
		if !strings.HasPrefix(url, known.prefix) {
			continue
		}

		path := strings.TrimPrefix(url, known.prefix)
		repoInfo, err = parseRepoPath(path)
		if err != nil {
			return nil, true, err
		}
		repoInfo.Type = known.rtype
		repoInfo.Host = known.host
		return repoInfo, true, nil
	}

	return nil, false, nil
}

// parseSelfHostedSSHURL handles self-hosted GitLab SSH URLs, e.g.
// "git@gitlab.example.com:owner/repo.git". GitLab is assumed for any
// non-public, non-HTTP(S) host since self-hosted GitHub Enterprise doesn't
// use this SSH URL shape.
func parseSelfHostedSSHURL(url string) (repoInfo *RepoInfo, matched bool, err error) {
	if !strings.Contains(url, "@") || !strings.Contains(url, ":") || strings.HasPrefix(url, "http") {
		return nil, false, nil
	}

	parts := strings.SplitN(url, "@", splitIntoTwoParts)
	if len(parts) != splitIntoTwoParts {
		return nil, false, nil
	}

	hostPathParts := strings.SplitN(parts[1], ":", splitIntoTwoParts)
	if len(hostPathParts) != splitIntoTwoParts {
		return nil, false, nil
	}

	host, path := hostPathParts[0], hostPathParts[1]

	repoInfo, err = parseRepoPath(path)
	if err != nil {
		return nil, true, err
	}
	repoInfo.Type = RepositoryTypeGitLab // Assume GitLab for self-hosted
	repoInfo.Host = host
	return repoInfo, true, nil
}

// parseSelfHostedHTTPURL handles self-hosted GitLab HTTPS/HTTP URLs, e.g.
// "https://gitlab.example.com/owner/repo.git". GitLab is assumed for any
// self-hosted host, matching parseSelfHostedSSHURL's assumption.
func parseSelfHostedHTTPURL(url string) (repoInfo *RepoInfo, matched bool, err error) {
	var rest string
	switch {
	case strings.HasPrefix(url, "https://"):
		rest = strings.TrimPrefix(url, "https://")
	case strings.HasPrefix(url, "http://"):
		rest = strings.TrimPrefix(url, "http://")
	default:
		return nil, false, nil
	}

	slashIndex := strings.Index(rest, "/")
	if slashIndex == -1 {
		return nil, true, fmt.Errorf("invalid repository URL format: %s", url)
	}

	host := rest[:slashIndex]
	path := rest[slashIndex+1:]

	repoInfo, err = parseRepoPath(path)
	if err != nil {
		return nil, true, err
	}
	repoInfo.Type = RepositoryTypeGitLab // Assume GitLab for self-hosted
	repoInfo.Host = host
	return repoInfo, true, nil
}

// parseGitHubURL extracts owner and repo name from various GitHub URL formats (kept for backward compatibility)
func parseGitHubURL(url string) (*RepoInfo, error) {
	repoInfo, err := parseRepositoryURL(url)
	if err != nil {
		return nil, err
	}
	if repoInfo.Type != RepositoryTypeGitHub {
		return nil, fmt.Errorf("not a GitHub repository: %s", url)
	}
	return repoInfo, nil
}

// parseRepoPath parses owner/repo from the path part of a GitHub URL
func parseRepoPath(path string) (*RepoInfo, error) {
	// Remove .git suffix if present
	path = strings.TrimSuffix(path, ".git")

	// Split by slash
	parts := strings.Split(path, "/")
	if len(parts) < splitIntoTwoParts {
		return nil, fmt.Errorf("invalid repository path: %s", path)
	}

	// Take first two parts as owner/repo
	return &RepoInfo{
		Owner: parts[0],
		Name:  parts[1],
	}, nil
}
