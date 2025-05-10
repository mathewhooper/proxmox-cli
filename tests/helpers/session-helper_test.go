package helpers_test

import (
	"os"
	"testing"

	"proxmox-cli/helpers"
)

func TestWriteSessionFile(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	sessionData := map[string]interface{}{
		"key": "value",
	}

	err := helpers.WriteSessionFile(sessionData)
	if err != nil {
		t.Errorf("Failed to write session file: %v", err)
	}
}

func TestReadSessionFile(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	sessionData := map[string]interface{}{
		"key": "value",
	}

	err := helpers.WriteSessionFile(sessionData)
	if err != nil {
		t.Fatalf("Failed to write session file: %v", err)
	}

	readData, err := helpers.ReadSessionFile()
	if err != nil {
		t.Errorf("Failed to read session file: %v", err)
	}

	if readData["key"] != "value" {
		t.Errorf("Expected 'value', got %v", readData["key"])
	}
}

func TestUpdateSessionField(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	sessionData := map[string]interface{}{
		"key": "value",
	}

	err := helpers.WriteSessionFile(sessionData)
	if err != nil {
		t.Fatalf("Failed to write session file: %v", err)
	}

	err = helpers.UpdateSessionField("key", "newValue")
	if err != nil {
		t.Errorf("Failed to update session field: %v", err)
	}

	updatedData, err := helpers.ReadSessionFile()
	if err != nil {
		t.Errorf("Failed to read session file: %v", err)
	}

	if updatedData["key"] != "newValue" {
		t.Errorf("Expected 'newValue', got %v", updatedData["key"])
	}
}
