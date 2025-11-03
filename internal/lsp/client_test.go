package lsp

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Use a temporary directory for testing
	rootPath := "/tmp"
	compileDBPath := "/tmp"

	client, err := NewClient(compileDBPath, rootPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Give it a moment to fully initialize
	time.Sleep(100 * time.Millisecond)

	// Verify client is working
	info, err := client.GetServerInfo()
	if err != nil {
		t.Errorf("GetServerInfo failed: %v", err)
	}

	if info == "" {
		t.Error("Expected non-empty server info")
	}

	t.Logf("✓ Client created successfully: %s", info)
}

func TestClientClose(t *testing.T) {
	rootPath := "/tmp"
	compileDBPath := "/tmp"

	client, err := NewClient(compileDBPath, rootPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test clean shutdown
	err = client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	t.Log("✓ Client closed successfully")
}
