package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	// Basic smoke test - just verify the root command is set up
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
	}
	if rootCmd.Use != "difftron" {
		t.Errorf("expected rootCmd.Use to be 'difftron', got %q", rootCmd.Use)
	}
}

func TestVersion(t *testing.T) {
	// Verify version variables are set
	if version == "" {
		t.Error("version should be set")
	}
	if commit == "" {
		t.Error("commit should be set")
	}
	if date == "" {
		t.Error("date should be set")
	}
}
