package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	
	configContent := `auth_url: "https://test.openstack.com/keystone/v3"
username: "testuser@example.com"
domain_name: "default"
project_id: "test-project-id"
project_name: "test-project"
region_name: "test-region"
availability_zone: "test-az"
image_name: "test-image"
flavor_name: "m1.small"
network_name: "test-network"
network_attachment_mode: "existing_network"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	
	config, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFromFile failed: %v", err)
	}
	
	if config.AuthURL != "https://test.openstack.com/keystone/v3" {
		t.Errorf("Expected AuthURL 'https://test.openstack.com/keystone/v3', got '%s'", config.AuthURL)
	}
	if config.Username != "testuser@example.com" {
		t.Errorf("Expected Username 'testuser@example.com', got '%s'", config.Username)
	}
	if config.DomainName != "default" {
		t.Errorf("Expected DomainName 'default', got '%s'", config.DomainName)
	}
	if config.ProjectID != "test-project-id" {
		t.Errorf("Expected ProjectID 'test-project-id', got '%s'", config.ProjectID)
	}
	if config.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName 'test-project', got '%s'", config.ProjectName)
	}
	if config.RegionName != "test-region" {
		t.Errorf("Expected RegionName 'test-region', got '%s'", config.RegionName)
	}
	if config.AvailabilityZone != "test-az" {
		t.Errorf("Expected AvailabilityZone 'test-az', got '%s'", config.AvailabilityZone)
	}
	if config.ImageName != "test-image" {
		t.Errorf("Expected ImageName 'test-image', got '%s'", config.ImageName)
	}
	if config.FlavorName != "m1.small" {
		t.Errorf("Expected FlavorName 'm1.small', got '%s'", config.FlavorName)
	}
	if config.NetworkName != "test-network" {
		t.Errorf("Expected NetworkName 'test-network', got '%s'", config.NetworkName)
	}
	if config.NetworkAttachmentMode != "existing_network" {
		t.Errorf("Expected NetworkAttachmentMode 'existing_network', got '%s'", config.NetworkAttachmentMode)
	}
}

func TestLoadConfigFromFile_InvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yaml")
	
	configContent := `invalid: yaml: content: [`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	
	_, err = loadConfigFromFile(configPath)
	if err == nil {
		t.Error("loadConfigFromFile should fail for invalid YAML")
	}
}

func TestLoadConfigFromFile_NonExistent(t *testing.T) {
	_, err := loadConfigFromFile("/non/existent/path/config.yaml")
	if err == nil {
		t.Error("loadConfigFromFile should fail for non-existent file")
	}
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Set up environment variables
	originalEnv := make(map[string]string)
	envVars := map[string]string{
		"OS_AUTH_URL":              "https://env.openstack.com/keystone/v3",
		"OS_USERNAME":               "envuser@example.com",
		"OS_DOMAIN_NAME":            "env-domain",
		"OS_PROJECT_ID":             "env-project-id",
		"OS_PROJECT_NAME":            "env-project",
		"OS_REGION_NAME":            "env-region",
		"OS_AVAILABILITY_ZONE":       "env-az",
		"OS_IMAGE_NAME":              "env-image",
		"OS_FLAVOR_NAME":             "env-flavor",
		"OS_NETWORK_NAME":            "env-network",
		"OS_NETWORK_ATTACHMENT_MODE": "env-mode",
		"OS_PASSWORD":                "test-password",
	}
	
	// Save original values and set new ones
	for key, value := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}
	
	// Restore original values after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	// Change to a temp directory where no config file exists
	tmpDir := t.TempDir()
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalCwd); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()
	
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	if config.AuthURL != "https://env.openstack.com/keystone/v3" {
		t.Errorf("Expected AuthURL from env, got '%s'", config.AuthURL)
	}
	if config.Username != "envuser@example.com" {
		t.Errorf("Expected Username from env, got '%s'", config.Username)
	}
	if config.DomainName != "env-domain" {
		t.Errorf("Expected DomainName from env, got '%s'", config.DomainName)
	}
	if config.Password != "test-password" {
		t.Errorf("Expected Password from env, got '%s'", config.Password)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Set up minimal environment variables
	originalEnv := make(map[string]string)
	envVars := map[string]string{
		"OS_AUTH_URL":        "https://test.openstack.com/keystone/v3",
		"OS_USERNAME":        "testuser@example.com",
		"OS_PROJECT_ID":       "test-project-id",
		"OS_PROJECT_NAME":     "test-project",
		"OS_REGION_NAME":      "test-region",
		"OS_AVAILABILITY_ZONE": "test-az",
		"OS_IMAGE_NAME":       "test-image",
		"OS_NETWORK_NAME":     "test-network",
		"OS_PASSWORD":         "test-password",
	}
	
	// Save original values and set new ones
	for key, value := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}
	
	// Restore original values after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	// Change to a temp directory where no config file exists
	tmpDir := t.TempDir()
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalCwd); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()
	
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	// Check defaults
	if config.DomainName != "default" {
		t.Errorf("Expected default DomainName 'default', got '%s'", config.DomainName)
	}
	if config.FlavorName != "m1.small" {
		t.Errorf("Expected default FlavorName 'm1.small', got '%s'", config.FlavorName)
	}
	if config.NetworkAttachmentMode != "existing_network" {
		t.Errorf("Expected default NetworkAttachmentMode 'existing_network', got '%s'", config.NetworkAttachmentMode)
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	// Set up minimal required environment variables but leave some out
	originalEnv := make(map[string]string)
	envVars := map[string]string{
		"OS_AUTH_URL":        "https://test.openstack.com/keystone/v3",
		"OS_USERNAME":        "testuser@example.com",
		"OS_PROJECT_ID":     "test-project-id",
		"OS_PROJECT_NAME":    "test-project",
		"OS_REGION_NAME":     "test-region",
		"OS_AVAILABILITY_ZONE": "test-az",
		"OS_IMAGE_NAME":     "test-image",
		"OS_NETWORK_NAME":    "test-network",
		"OS_PASSWORD":        "test-password",
	}
	
	// Save original values and set new ones
	for key, value := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}
	
	// Restore original values after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	// Change to a temp directory where no config file exists
	tmpDir := t.TempDir()
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalCwd); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()
	
	// Test missing OS_AUTH_URL
	os.Unsetenv("OS_AUTH_URL")
	_, err = LoadConfig()
	if err == nil {
		t.Error("LoadConfig should fail when OS_AUTH_URL is missing")
	}
	
	// Test missing OS_USERNAME
	os.Setenv("OS_AUTH_URL", "https://test.openstack.com/keystone/v3")
	os.Unsetenv("OS_USERNAME")
	_, err = LoadConfig()
	if err == nil {
		t.Error("LoadConfig should fail when OS_USERNAME is missing")
	}
	
	// Test missing OS_PROJECT_ID
	os.Setenv("OS_USERNAME", "testuser@example.com")
	os.Unsetenv("OS_PROJECT_ID")
	_, err = LoadConfig()
	if err == nil {
		t.Error("LoadConfig should fail when OS_PROJECT_ID is missing")
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	// Test with existing env var
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")
	
	result := getEnvWithDefault("TEST_VAR", "default-value")
	if result != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", result)
	}
	
	// Test with non-existent env var
	result = getEnvWithDefault("NON_EXISTENT_VAR", "default-value")
	if result != "default-value" {
		t.Errorf("Expected 'default-value', got '%s'", result)
	}
}
