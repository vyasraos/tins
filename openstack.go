package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
)

// OpenStackClient wraps the OpenStack clients
type OpenStackClient struct {
	computeClient *gophercloud.ServiceClient
	networkClient *gophercloud.ServiceClient
	imageClient   *gophercloud.ServiceClient
	config        *OpenStackConfig
}

// NewOpenStackClient creates a new OpenStack client
func NewOpenStackClient(ctx context.Context, config *OpenStackConfig) (*OpenStackClient, error) {
	// Authenticate with OpenStack
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: config.AuthURL,
		Username:         config.Username,
		Password:         config.Password,
		DomainName:       config.DomainName,
		TenantID:         config.ProjectID,
	}

	provider, err := openstack.AuthenticatedClient(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Initialize Compute service client
	computeClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: config.RegionName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	// Initialize Network service client
	networkClient, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Region: config.RegionName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create network client: %w", err)
	}

	// Initialize Image service client (Glance)
	imageClient, err := openstack.NewImageV2(provider, gophercloud.EndpointOpts{
		Region: config.RegionName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create image client: %w", err)
	}

	return &OpenStackClient{
		computeClient: computeClient,
		networkClient: networkClient,
		imageClient:   imageClient,
		config:        config,
	}, nil
}

// FindImageByName finds an image by name
func (c *OpenStackClient) FindImageByName(ctx context.Context, imageName string) (string, error) {
	listOpts := images.ListOpts{
		Name: imageName,
	}
	allPages, err := images.List(c.imageClient, listOpts).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list images: %w", err)
	}

	allImages, err := images.ExtractImages(allPages)
	if err != nil {
		return "", fmt.Errorf("failed to extract images: %w", err)
	}

	if len(allImages) == 0 {
		return "", fmt.Errorf("image '%s' not found", imageName)
	}

	// Return the first matching image ID
	return allImages[0].ID, nil
}

// FindFlavorByName finds a flavor by name
func (c *OpenStackClient) FindFlavorByName(ctx context.Context, flavorName string) (string, error) {
	listOpts := flavors.ListOpts{
		AccessType: flavors.PublicAccess,
	}
	allPages, err := flavors.ListDetail(c.computeClient, listOpts).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list flavors: %w", err)
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return "", fmt.Errorf("failed to extract flavors: %w", err)
	}

	for _, flavor := range allFlavors {
		if flavor.Name == flavorName {
			return flavor.ID, nil
		}
	}

	return "", fmt.Errorf("flavor '%s' not found", flavorName)
}

// FindNetworkByName finds a network by name
func (c *OpenStackClient) FindNetworkByName(ctx context.Context, networkName string) (string, error) {
	listOpts := networks.ListOpts{
		Name: networkName,
	}
	allPages, err := networks.List(c.networkClient, listOpts).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}

	allNetworks, err := networks.ExtractNetworks(allPages)
	if err != nil {
		return "", fmt.Errorf("failed to extract networks: %w", err)
	}

	if len(allNetworks) == 0 {
		return "", fmt.Errorf("network '%s' not found", networkName)
	}

	return allNetworks[0].ID, nil
}

// CreateInstance creates a new temporary instance
func (c *OpenStackClient) CreateInstance(ctx context.Context, instanceName string, publicKey string) (*servers.Server, error) {
	// Find image ID
	imageID, err := c.FindImageByName(ctx, c.config.ImageName)
	if err != nil {
		return nil, err
	}

	// Find flavor ID
	flavorID, err := c.FindFlavorByName(ctx, c.config.FlavorName)
	if err != nil {
		return nil, err
	}

	// Find network ID
	networkID, err := c.FindNetworkByName(ctx, c.config.NetworkName)
	if err != nil {
		return nil, err
	}

	// Create server options
	createOpts := servers.CreateOpts{
		Name:      instanceName,
		ImageRef:  imageID,
		FlavorRef: flavorID,
		Networks: []servers.Network{
			{UUID: networkID},
		},
		AvailabilityZone: c.config.AvailabilityZone,
		Metadata: map[string]string{
			TempInstanceTag: "true",
		},
		UserData: []byte(fmt.Sprintf(`#!/bin/bash
# Add public key to authorized_keys
mkdir -p /root/.ssh
echo "%s" >> /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys
chmod 700 /root/.ssh
`, publicKey)),
	}

	// Note: Tags are added after instance creation via separate API call
	// as some OpenStack versions don't support tags in CreateOpts

	// Create the server
	server, err := servers.Create(ctx, c.computeClient, createOpts, nil).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return server, nil
}

// ListInstances lists all temporary instances
func (c *OpenStackClient) ListInstances(ctx context.Context) ([]servers.Server, error) {
	// List all servers and filter by metadata since tags aren't supported in all OpenStack versions
	listOpts := servers.ListOpts{}
	allPages, err := servers.List(c.computeClient, listOpts).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract servers: %w", err)
	}

	// Filter by tins metadata and name prefix
	var tinsInstances []servers.Server
	for _, server := range allServers {
		// Check if it has the tins metadata or starts with tins- prefix
		if server.Metadata != nil {
			if val, ok := server.Metadata[TempInstanceTag]; ok && val == "true" {
				tinsInstances = append(tinsInstances, server)
				continue
			}
		}
		// Also check by name prefix as fallback
		if len(server.Name) >= len(InstanceNamePrefix) && server.Name[:len(InstanceNamePrefix)] == InstanceNamePrefix {
			tinsInstances = append(tinsInstances, server)
		}
	}

	return tinsInstances, nil
}

// GetInstance retrieves a server by ID
func (c *OpenStackClient) GetInstance(ctx context.Context, serverID string) (*servers.Server, error) {
	server, err := servers.Get(ctx, c.computeClient, serverID).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}
	return server, nil
}

// DeleteInstance deletes a server
func (c *OpenStackClient) DeleteInstance(ctx context.Context, serverID string) error {
	err := servers.Delete(ctx, c.computeClient, serverID).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}
	return nil
}

// WaitForInstanceActive waits for an instance to become active
func (c *OpenStackClient) WaitForInstanceActive(ctx context.Context, serverID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			server, err := c.GetInstance(ctx, serverID)
			if err != nil {
				return err
			}
			if server.Status == "ACTIVE" {
				return nil
			}
			if server.Status == "ERROR" {
				return fmt.Errorf("server entered ERROR state")
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for server to become active")
			}
		}
	}
}
