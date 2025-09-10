package main

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNotGitRepo = errors.New("not a git repository")

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