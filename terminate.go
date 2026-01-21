package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var terminateCmd = &cobra.Command{
	Use:   "terminate [instance-name-or-id]",
	Short: "Terminate a temporary instance",
	Long:  "Terminate an ephemeral OpenStack instance and delete its associated SSH keys.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceIdentifier := args[0]

		// Load configuration
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		// Create OpenStack client
		ctx := context.Background()
		client, err := NewOpenStackClient(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to create OpenStack client: %w", err)
		}

		var serverID string
		var instanceName string

		// Check if it's an instance name or ID
		if strings.HasPrefix(instanceIdentifier, InstanceNamePrefix) {
			// It's a full instance name
			instanceName = strings.TrimPrefix(instanceIdentifier, InstanceNamePrefix)
			// Find the server by name
			servers, err := client.ListInstances(ctx)
			if err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}

			found := false
			for _, s := range servers {
				if s.Name == instanceIdentifier {
					serverID = s.ID
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("instance '%s' not found", instanceIdentifier)
			}
		} else if strings.HasPrefix(instanceIdentifier, InstanceNamePrefix) {
			// It's just the name part
			instanceName = instanceIdentifier
			// Find the server by name
			servers, err := client.ListInstances(ctx)
			if err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}

			fullName := fmt.Sprintf("%s%s", InstanceNamePrefix, instanceIdentifier)
			found := false
			for _, s := range servers {
				if s.Name == fullName {
					serverID = s.ID
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("instance '%s' not found", instanceIdentifier)
			}
		} else {
			// Assume it's an instance ID
			serverID = instanceIdentifier
			server, err := client.GetInstance(ctx, serverID)
			if err != nil {
				return fmt.Errorf("failed to get instance: %w", err)
			}
			// Extract instance name from full name
			if strings.HasPrefix(server.Name, InstanceNamePrefix) {
				instanceName = strings.TrimPrefix(server.Name, InstanceNamePrefix)
			} else {
				// If it doesn't have the prefix, try to find it in the list
				servers, err := client.ListInstances(ctx)
				if err == nil {
					for _, s := range servers {
						if s.ID == serverID {
							if strings.HasPrefix(s.Name, InstanceNamePrefix) {
								instanceName = strings.TrimPrefix(s.Name, InstanceNamePrefix)
							}
							break
						}
					}
				}
				if instanceName == "" {
					instanceName = server.Name
				}
			}
		}

		// Delete the instance
		fmt.Printf("Terminating instance %s (ID: %s)...\n", instanceIdentifier, serverID)
		if err := client.DeleteInstance(ctx, serverID); err != nil {
			return fmt.Errorf("failed to delete instance: %w", err)
		}
		fmt.Printf("Instance terminated successfully.\n")

		// Delete SSH keys - always try to clean up if we have an instance name
		if instanceName != "" {
			fmt.Printf("Cleaning up SSH keys for %s...\n", instanceName)
			if err := DeleteSSHKey(instanceName); err != nil {
				// Don't fail if keys don't exist, just warn
				fmt.Printf("Warning: Failed to delete SSH keys (they may not exist): %v\n", err)
			} else {
				fmt.Printf("SSH keys deleted successfully.\n")
			}
		} else {
			fmt.Printf("Warning: Could not determine instance name, skipping SSH key cleanup\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(terminateCmd)
}
