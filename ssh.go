package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// SSHKeyPair represents an SSH key pair
type SSHKeyPair struct {
	PrivateKeyPath string
	PublicKeyPath  string
	PublicKey      string
}

// GenerateSSHKey generates a new SSH key pair
func GenerateSSHKey(instanceName string) (*SSHKeyPair, error) {
	// Create ~/.ssh directory if it doesn't exist
	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Encode private key to PEM
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privateKeyPath := filepath.Join(sshDir, fmt.Sprintf("%s%s", InstanceNamePrefix, instanceName))
	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(privateKeyPEM), 0600); err != nil {
		return nil, fmt.Errorf("failed to write private key: %w", err)
	}

	// Generate public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	publicKeyString := string(ssh.MarshalAuthorizedKey(publicKey))
	publicKeyPath := fmt.Sprintf("%s.pub", privateKeyPath)
	if err := os.WriteFile(publicKeyPath, []byte(publicKeyString), 0644); err != nil {
		return nil, fmt.Errorf("failed to write public key: %w", err)
	}

	return &SSHKeyPair{
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		PublicKey:      publicKeyString,
	}, nil
}

// DeleteSSHKey deletes the SSH key pair for an instance
func DeleteSSHKey(instanceName string) error {
	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	privateKeyPath := filepath.Join(sshDir, fmt.Sprintf("%s%s", InstanceNamePrefix, instanceName))
	publicKeyPath := fmt.Sprintf("%s.pub", privateKeyPath)

	// Delete private key
	if err := os.Remove(privateKeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete private key: %w", err)
	}

	// Delete public key
	if err := os.Remove(publicKeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete public key: %w", err)
	}

	return nil
}

// GetSSHKeyPath returns the SSH key path for an instance
func GetSSHKeyPath(instanceName string) string {
	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	return filepath.Join(sshDir, fmt.Sprintf("%s%s", InstanceNamePrefix, instanceName))
}
