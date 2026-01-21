package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSSHKey(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tmpDir)
	
	instanceName := "test-instance"
	
	keyPair, err := GenerateSSHKey(instanceName)
	if err != nil {
		t.Fatalf("GenerateSSHKey failed: %v", err)
	}
	
	// Verify key pair structure
	if keyPair == nil {
		t.Fatal("GenerateSSHKey returned nil key pair")
	}
	
	// Check that private key path is correct
	expectedPrivatePath := filepath.Join(tmpDir, ".ssh", InstanceNamePrefix+instanceName)
	if keyPair.PrivateKeyPath != expectedPrivatePath {
		t.Errorf("Expected private key path %s, got %s", expectedPrivatePath, keyPair.PrivateKeyPath)
	}
	
	// Check that public key path is correct
	expectedPublicPath := expectedPrivatePath + ".pub"
	if keyPair.PublicKeyPath != expectedPublicPath {
		t.Errorf("Expected public key path %s, got %s", expectedPublicPath, keyPair.PublicKeyPath)
	}
	
	// Verify files exist
	if _, err := os.Stat(keyPair.PrivateKeyPath); os.IsNotExist(err) {
		t.Errorf("Private key file does not exist: %s", keyPair.PrivateKeyPath)
	}
	
	if _, err := os.Stat(keyPair.PublicKeyPath); os.IsNotExist(err) {
		t.Errorf("Public key file does not exist: %s", keyPair.PublicKeyPath)
	}
	
	// Verify file permissions
	privateKeyInfo, err := os.Stat(keyPair.PrivateKeyPath)
	if err != nil {
		t.Fatalf("Failed to stat private key: %v", err)
	}
	if privateKeyInfo.Mode().Perm() != 0600 {
		t.Errorf("Expected private key permissions 0600, got %o", privateKeyInfo.Mode().Perm())
	}
	
	publicKeyInfo, err := os.Stat(keyPair.PublicKeyPath)
	if err != nil {
		t.Fatalf("Failed to stat public key: %v", err)
	}
	if publicKeyInfo.Mode().Perm() != 0644 {
		t.Errorf("Expected public key permissions 0644, got %o", publicKeyInfo.Mode().Perm())
	}
	
	// Verify public key format (should start with ssh-rsa)
	if len(keyPair.PublicKey) == 0 {
		t.Error("Public key is empty")
	}
	if keyPair.PublicKey[:7] != "ssh-rsa" {
		t.Errorf("Expected public key to start with 'ssh-rsa', got: %s", keyPair.PublicKey[:min(20, len(keyPair.PublicKey))])
	}
}

func TestDeleteSSHKey(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tmpDir)
	
	instanceName := "test-instance"
	
	// First, generate a key pair
	keyPair, err := GenerateSSHKey(instanceName)
	if err != nil {
		t.Fatalf("GenerateSSHKey failed: %v", err)
	}
	
	// Verify files exist before deletion
	if _, err := os.Stat(keyPair.PrivateKeyPath); os.IsNotExist(err) {
		t.Fatal("Private key file should exist before deletion")
	}
	if _, err := os.Stat(keyPair.PublicKeyPath); os.IsNotExist(err) {
		t.Fatal("Public key file should exist before deletion")
	}
	
	// Delete the keys
	err = DeleteSSHKey(instanceName)
	if err != nil {
		t.Fatalf("DeleteSSHKey failed: %v", err)
	}
	
	// Verify files are deleted
	if _, err := os.Stat(keyPair.PrivateKeyPath); !os.IsNotExist(err) {
		t.Error("Private key file should be deleted")
	}
	if _, err := os.Stat(keyPair.PublicKeyPath); !os.IsNotExist(err) {
		t.Error("Public key file should be deleted")
	}
}

func TestDeleteSSHKey_NonExistent(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tmpDir)
	
	instanceName := "non-existent-instance"
	
	// Deleting non-existent keys should not error
	err := DeleteSSHKey(instanceName)
	if err != nil {
		t.Errorf("DeleteSSHKey should not error for non-existent keys, got: %v", err)
	}
}

func TestGetSSHKeyPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tmpDir)
	
	instanceName := "test-instance"
	expectedPath := filepath.Join(tmpDir, ".ssh", InstanceNamePrefix+instanceName)
	
	path := GetSSHKeyPath(instanceName)
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
