package main

import (
	"os"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Test that version command exists and can be executed
	rootCmd.SetArgs([]string{"version"})
	
	// Capture output
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	
	// We can't easily test the output without refactoring, but we can test
	// that the command doesn't crash
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Version command failed: %v", err)
	}
	
	// Restore
	os.Stdout = originalStdout
	os.Stderr = originalStderr
}

func TestConstants(t *testing.T) {
	if InstanceNamePrefix != "tins-" {
		t.Errorf("Expected InstanceNamePrefix 'tins-', got '%s'", InstanceNamePrefix)
	}
	
	if TempInstanceTag != "tins" {
		t.Errorf("Expected TempInstanceTag 'tins', got '%s'", TempInstanceTag)
	}
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables are set (default values)
	if Version == "" {
		t.Error("Version should not be empty (should default to 'dev')")
	}
	
	if Commit == "" {
		t.Error("Commit should not be empty (should default to 'unknown')")
	}
}
