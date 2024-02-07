package main

import (
	"os"
	"testing"
)

func TestGetLogFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "testLogFile")
	if err != nil {
		t.Fatalf("error while opening test file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	t.Run("success get log file", func(t *testing.T) {
		file, err := GetLogFile(tempFile.Name())
		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
		if file == nil {
			t.Errorf("expected file: %v, got: %v", file, err)
		}
	})

	t.Run("failed to get log file", func(t *testing.T) {
		_, err := GetLogFile("/some/wrong/path")
		if err == nil {
			t.Errorf("expected error: %v, got nil", err)
		}
	})
}
