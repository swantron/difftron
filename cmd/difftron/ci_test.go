package main

import (
	"os"
	"testing"
)

func TestIsValidSHA(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid full SHA", "abc123def4567890123456789012345678901234", true},
		{"valid short SHA", "abc1234", true},
		{"too short", "abc123", false},
		{"too long", "abc123def45678901234567890123456789012345", false},
		{"invalid chars", "abc123g", false},
		{"valid hex uppercase", "ABC1234", true},
		{"mixed case", "AbC1234", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSHA(tt.input)
			if result != tt.expected {
				t.Errorf("isValidSHA(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid hex lowercase", "abc123", true},
		{"valid hex uppercase", "ABC123", true},
		{"mixed case", "AbC123", true},
		{"invalid char", "abc123g", false},
		{"numbers only", "123456", true},
		{"letters only", "abcdef", true},
		{"empty", "", true},
		{"special chars", "abc-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHex(tt.input)
			if result != tt.expected {
				t.Errorf("isHex(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectBaseRef(t *testing.T) {
	// Save original env
	originalBaseRef := os.Getenv("GITHUB_BASE_REF")
	originalBaseSHA := os.Getenv("GITHUB_BASE_SHA")
	originalEventBefore := os.Getenv("GITHUB_EVENT_BEFORE")
	originalMRBaseSHA := os.Getenv("CI_MERGE_REQUEST_DIFF_BASE_SHA")

	defer func() {
		os.Setenv("GITHUB_BASE_REF", originalBaseRef)
		os.Setenv("GITHUB_BASE_SHA", originalBaseSHA)
		os.Setenv("GITHUB_EVENT_BEFORE", originalEventBefore)
		os.Setenv("CI_MERGE_REQUEST_DIFF_BASE_SHA", originalMRBaseSHA)
	}()

	t.Run("GitHub Actions with base SHA", func(t *testing.T) {
		os.Setenv("GITHUB_BASE_REF", "main")
		os.Setenv("GITHUB_BASE_SHA", "abc1234")
		os.Unsetenv("GITHUB_EVENT_BEFORE")
		os.Unsetenv("CI_MERGE_REQUEST_DIFF_BASE_SHA")

		result := detectBaseRef()
		if result != "abc1234" {
			t.Errorf("expected 'abc1234', got %q", result)
		}
	})

	t.Run("GitHub Actions with event before", func(t *testing.T) {
		os.Unsetenv("GITHUB_BASE_REF")
		os.Unsetenv("GITHUB_BASE_SHA")
		os.Setenv("GITHUB_EVENT_BEFORE", "def5678")
		os.Unsetenv("CI_MERGE_REQUEST_DIFF_BASE_SHA")

		result := detectBaseRef()
		if result != "def5678" {
			t.Errorf("expected 'def5678', got %q", result)
		}
	})

	t.Run("GitLab CI", func(t *testing.T) {
		os.Unsetenv("GITHUB_BASE_REF")
		os.Unsetenv("GITHUB_BASE_SHA")
		os.Unsetenv("GITHUB_EVENT_BEFORE")
		os.Setenv("CI_MERGE_REQUEST_DIFF_BASE_SHA", "ghi9012")

		result := detectBaseRef()
		if result != "ghi9012" {
			t.Errorf("expected 'ghi9012', got %q", result)
		}
	})

	t.Run("default fallback", func(t *testing.T) {
		os.Unsetenv("GITHUB_BASE_REF")
		os.Unsetenv("GITHUB_BASE_SHA")
		os.Unsetenv("GITHUB_EVENT_BEFORE")
		os.Unsetenv("CI_MERGE_REQUEST_DIFF_BASE_SHA")

		result := detectBaseRef()
		if result != "HEAD~1" {
			t.Errorf("expected 'HEAD~1', got %q", result)
		}
	})
}

func TestDetectHeadRef(t *testing.T) {
	// Save original env
	originalHeadSHA := os.Getenv("GITHUB_HEAD_SHA")
	originalSHA := os.Getenv("GITHUB_SHA")
	originalCommitSHA := os.Getenv("CI_COMMIT_SHA")

	defer func() {
		os.Setenv("GITHUB_HEAD_SHA", originalHeadSHA)
		os.Setenv("GITHUB_SHA", originalSHA)
		os.Setenv("CI_COMMIT_SHA", originalCommitSHA)
	}()

	t.Run("GitHub Actions with head SHA", func(t *testing.T) {
		os.Setenv("GITHUB_HEAD_SHA", "abc1234")
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("CI_COMMIT_SHA")

		result := detectHeadRef()
		if result != "abc1234" {
			t.Errorf("expected 'abc1234', got %q", result)
		}
	})

	t.Run("GitHub Actions with SHA", func(t *testing.T) {
		os.Unsetenv("GITHUB_HEAD_SHA")
		os.Setenv("GITHUB_SHA", "def5678")
		os.Unsetenv("CI_COMMIT_SHA")

		result := detectHeadRef()
		if result != "def5678" {
			t.Errorf("expected 'def5678', got %q", result)
		}
	})

	t.Run("GitLab CI", func(t *testing.T) {
		os.Unsetenv("GITHUB_HEAD_SHA")
		os.Unsetenv("GITHUB_SHA")
		os.Setenv("CI_COMMIT_SHA", "ghi9012")

		result := detectHeadRef()
		if result != "ghi9012" {
			t.Errorf("expected 'ghi9012', got %q", result)
		}
	})

	t.Run("default fallback", func(t *testing.T) {
		os.Unsetenv("GITHUB_HEAD_SHA")
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("CI_COMMIT_SHA")

		result := detectHeadRef()
		if result != "HEAD" {
			t.Errorf("expected 'HEAD', got %q", result)
		}
	})
}
