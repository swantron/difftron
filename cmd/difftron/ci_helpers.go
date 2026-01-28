package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// detectBaseRef detects the base git ref from CI environment variables
func detectBaseRef() string {
	// GitHub Actions
	if base := os.Getenv("GITHUB_BASE_REF"); base != "" {
		// For PRs, get the base SHA
		if sha := os.Getenv("GITHUB_BASE_SHA"); sha != "" {
			return sha
		}
		return base
	}
	if before := os.Getenv("GITHUB_EVENT_BEFORE"); before != "" {
		return before
	}

	// GitLab CI
	if base := os.Getenv("CI_MERGE_REQUEST_DIFF_BASE_SHA"); base != "" {
		return base
	}

	// Default: compare against previous commit
	return "HEAD~1"
}

// detectHeadRef detects the head git ref from CI environment variables
func detectHeadRef() string {
	// GitHub Actions
	if head := os.Getenv("GITHUB_HEAD_SHA"); head != "" {
		return head
	}
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		return sha
	}

	// GitLab CI
	if head := os.Getenv("CI_COMMIT_SHA"); head != "" {
		return head
	}

	// Default: current HEAD
	return "HEAD"
}

// getGitDiffForPR gets git diff optimized for PRs (uses three-dot merge-base)
func getGitDiffForPR(base, head string) (string, error) {
	// Handle special case where base might be a branch name
	if !strings.HasPrefix(base, "HEAD") && !isValidSHA(base) {
		// Try to get the SHA of the base branch
		cmd := exec.Command("git", "rev-parse", base)
		baseSHA, err := cmd.Output()
		if err == nil {
			base = strings.TrimSpace(string(baseSHA))
		}
	}

	// Use three dots (...) for merge-base diff (better for PRs)
	// Falls back to two dots (..) if merge-base fails
	cmd := exec.Command("git", "diff", base+"..."+head)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to two-dot diff if three-dot fails
		cmd = exec.Command("git", "diff", base+".."+head)
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("git diff failed: %w", err)
		}
	}
	return string(output), nil
}

// isValidSHA checks if a string looks like a valid git SHA
func isValidSHA(s string) bool {
	// Basic SHA validation (40 chars for full SHA, 7+ for short)
	return len(s) >= 7 && len(s) <= 40 && isHex(s)
}

// isHex checks if a string contains only hexadecimal characters
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
