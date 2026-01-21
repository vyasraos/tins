package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFile represents the YAML configuration file structure
type ConfigFile struct {
	AuthURL    string `yaml:"auth_url"`
	Username   string `yaml:"username"`
	DomainName string `yaml:"domain_name"`

	ProjectID   string `yaml:"project_id"`
	ProjectName string `yaml:"project_name"`

	RegionName       string `yaml:"region_name"`
	AvailabilityZone string `yaml:"availability_zone"`

	ImageName  string `yaml:"image_name"`
	FlavorName string `yaml:"flavor_name"`

	NetworkName           string `yaml:"network_name"`
	NetworkAttachmentMode string `yaml:"network_attachment_mode"`
}

// OpenStackConfig holds the OpenStack-specific configuration loaded from YAML file and environment variables.
// This is provider-specific configuration for OpenStack. Future providers (AWS, GCP) will have
// their own config types (e.g., AWSConfig, GCPConfig).
type OpenStackConfig struct {
	// Authentication
	AuthURL    string // OpenStack Keystone authentication URL
	Username   string // OpenStack username
	Password   string // OpenStack password (must come from env var)
	DomainName string // OpenStack domain name (default: "default")

	// Project/Tenant
	ProjectID   string // OpenStack project ID
	ProjectName string // OpenStack project name

	// Region and Availability
	RegionName       string // OpenStack region name
	AvailabilityZone string // Availability zone for instances

	// Instance Configuration
	ImageName  string // Name of the image to use
	FlavorName string // Instance flavor name (default: "m1.small")

	// Network Configuration
	NetworkName           string // Name of the network to attach to
	NetworkAttachmentMode string // Network attachment mode (default: "existing_network")
}

// findConfigFile looks for the config file in the following locations:
// 1. .config/tint.yaml (current directory)
// 2. ~/.config/tint/tint.yaml (home directory)
func findConfigFile() (string, error) {
	// Check current directory
	cwd, err := os.Getwd()
	if err == nil {
		localConfig := filepath.Join(cwd, ".config", "tint.yaml")
		if _, err := os.Stat(localConfig); err == nil {
			return localConfig, nil
		}
	}

	// Check home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeConfig := filepath.Join(homeDir, ".config", "tint", "tint.yaml")
		if _, err := os.Stat(homeConfig); err == nil {
			return homeConfig, nil
		}
	}

	return "", fmt.Errorf("config file not found in .config/tint.yaml or ~/.config/tint/tint.yaml")
}

// loadConfigFromFile loads configuration from YAML file
func loadConfigFromFile(filePath string) (*ConfigFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadConfig loads OpenStack configuration from YAML file and environment variables.
// Environment variables override values from the config file.
// Password must be provided via OS_PASSWORD environment variable.
// Config file is searched in:
//  1. .config/tint.yaml (current directory)
//  2. ~/.config/tint/tint.yaml (home directory)
func LoadConfig() (*OpenStackConfig, error) {
	config := &OpenStackConfig{}

	// Try to load from config file
	configFile, err := findConfigFile()
	if err == nil {
		fileConfig, err := loadConfigFromFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile, err)
		}
		// Load non-sensitive values from file
		config.AuthURL = fileConfig.AuthURL
		config.Username = fileConfig.Username
		config.DomainName = fileConfig.DomainName
		config.ProjectID = fileConfig.ProjectID
		config.ProjectName = fileConfig.ProjectName
		config.RegionName = fileConfig.RegionName
		config.AvailabilityZone = fileConfig.AvailabilityZone
		config.ImageName = fileConfig.ImageName
		config.FlavorName = fileConfig.FlavorName
		config.NetworkName = fileConfig.NetworkName
		config.NetworkAttachmentMode = fileConfig.NetworkAttachmentMode
	}

	// Environment variables override config file values
	if authURL := os.Getenv("OS_AUTH_URL"); authURL != "" {
		config.AuthURL = authURL
	}
	if username := os.Getenv("OS_USERNAME"); username != "" {
		config.Username = username
	}
	if domainName := os.Getenv("OS_DOMAIN_NAME"); domainName != "" {
		config.DomainName = domainName
	}
	if projectID := os.Getenv("OS_PROJECT_ID"); projectID != "" {
		config.ProjectID = projectID
	}
	if projectName := os.Getenv("OS_PROJECT_NAME"); projectName != "" {
		config.ProjectName = projectName
	}
	if regionName := os.Getenv("OS_REGION_NAME"); regionName != "" {
		config.RegionName = regionName
	}
	if availabilityZone := os.Getenv("OS_AVAILABILITY_ZONE"); availabilityZone != "" {
		config.AvailabilityZone = availabilityZone
	}
	if imageName := os.Getenv("OS_IMAGE_NAME"); imageName != "" {
		config.ImageName = imageName
	}
	if flavorName := os.Getenv("OS_FLAVOR_NAME"); flavorName != "" {
		config.FlavorName = flavorName
	}
	if networkName := os.Getenv("OS_NETWORK_NAME"); networkName != "" {
		config.NetworkName = networkName
	}
	if networkAttachmentMode := os.Getenv("OS_NETWORK_ATTACHMENT_MODE"); networkAttachmentMode != "" {
		config.NetworkAttachmentMode = networkAttachmentMode
	}

	// Password must come from environment variable (sensitive)
	config.Password = getEnvRequired("OS_PASSWORD")

	// Set defaults for optional fields
	if config.DomainName == "" {
		config.DomainName = "default"
	}
	if config.FlavorName == "" {
		config.FlavorName = "m1.small"
	}
	if config.NetworkAttachmentMode == "" {
		config.NetworkAttachmentMode = "existing_network"
	}

	// Validate required fields
	if config.AuthURL == "" {
		return nil, fmt.Errorf("OS_AUTH_URL is required (set in config file or environment variable)")
	}
	if config.Username == "" {
		return nil, fmt.Errorf("OS_USERNAME is required (set in config file or environment variable)")
	}
	if config.ProjectID == "" {
		return nil, fmt.Errorf("OS_PROJECT_ID is required (set in config file or environment variable)")
	}
	if config.ProjectName == "" {
		return nil, fmt.Errorf("OS_PROJECT_NAME is required (set in config file or environment variable)")
	}
	if config.RegionName == "" {
		return nil, fmt.Errorf("OS_REGION_NAME is required (set in config file or environment variable)")
	}
	if config.AvailabilityZone == "" {
		return nil, fmt.Errorf("OS_AVAILABILITY_ZONE is required (set in config file or environment variable)")
	}
	if config.ImageName == "" {
		return nil, fmt.Errorf("OS_IMAGE_NAME is required (set in config file or environment variable)")
	}
	if config.NetworkName == "" {
		return nil, fmt.Errorf("OS_NETWORK_NAME is required (set in config file or environment variable)")
	}

	return config, nil
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Fprintf(os.Stderr, "Error: Required environment variable %s is not set\n", key)
		os.Exit(1)
	}
	return value
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
