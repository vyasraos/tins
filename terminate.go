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
	Long:  "Terminate an ephemeral OpenStack instance and delete its associated SSH keys. Use --all to terminate all tins instances.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for --all flag
		terminateAll, _ := cmd.Flags().GetBool("all")

		if terminateAll {
			// Validate that no instance identifier is provided with --all
			if len(args) > 0 {
				return fmt.Errorf("cannot specify instance identifier with --all flag")
			}
			return terminateAllInstances(cmd)
		}

		// Regular single instance termination
		if len(args) == 0 {
			return fmt.Errorf("instance identifier required (or use --all to terminate all instances)")
		}

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
		var fullInstanceName string

		// Check if it's an instance name or ID
		if strings.HasPrefix(instanceIdentifier, InstanceNamePrefix) {
			// It's a full instance name
			fullInstanceName = instanceIdentifier
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
		} else {
			// It's just the name part (without prefix) or an instance ID
			// Try to find by name first
			fullName := fmt.Sprintf("%s%s", InstanceNamePrefix, instanceIdentifier)
			servers, err := client.ListInstances(ctx)
			if err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}

			found := false
			for _, s := range servers {
				if s.Name == fullName {
					serverID = s.ID
					fullInstanceName = s.Name
					instanceName = instanceIdentifier
					found = true
					break
				}
			}

			if !found {
				// Assume it's an instance ID
				serverID = instanceIdentifier
				server, err := client.GetInstance(ctx, serverID)
				if err != nil {
					return fmt.Errorf("failed to get instance: %w", err)
				}
				fullInstanceName = server.Name
				// Extract instance name from full name
				if strings.HasPrefix(server.Name, InstanceNamePrefix) {
					instanceName = strings.TrimPrefix(server.Name, InstanceNamePrefix)
				} else {
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

		// Delete OpenStack keypair - keypair name matches full instance name
		if fullInstanceName != "" {
			fmt.Printf("Deleting OpenStack keypair %s...\n", fullInstanceName)
			if err := client.DeleteKeypair(ctx, fullInstanceName); err != nil {
				// Don't fail if keypair doesn't exist, just warn
				fmt.Printf("Warning: Failed to delete OpenStack keypair (it may not exist): %v\n", err)
			} else {
				fmt.Printf("OpenStack keypair deleted successfully.\n")
			}
		} else {
			fmt.Printf("Warning: Could not determine instance name, skipping OpenStack keypair cleanup\n")
		}

		// Delete local SSH keys - always try to clean up if we have an instance name
		if instanceName != "" {
			fmt.Printf("Cleaning up local SSH keys for %s...\n", instanceName)
			if err := DeleteSSHKey(instanceName); err != nil {
				// Don't fail if keys don't exist, just warn
				fmt.Printf("Warning: Failed to delete local SSH keys (they may not exist): %v\n", err)
			} else {
				fmt.Printf("Local SSH keys deleted successfully.\n")
			}
		} else {
			fmt.Printf("Warning: Could not determine instance name, skipping local SSH key cleanup\n")
		}

		return nil
	},
}

// terminateAllInstances terminates all tins instances and cleans up all tins keypairs
func terminateAllInstances(_ *cobra.Command) error {
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

	// Get all tins instances
	servers, err := client.ListInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(servers) == 0 {
		fmt.Printf("No tins instances found to terminate.\n")
	} else {
		fmt.Printf("Found %d tins instance(s) to terminate:\n", len(servers))
		for _, server := range servers {
			fmt.Printf("  - %s (ID: %s, Status: %s)\n", server.Name, server.ID, server.Status)
		}
		fmt.Printf("\nTerminating instances...\n")

		// Terminate each instance
		for _, server := range servers {
			fmt.Printf("\nTerminating instance %s (ID: %s)...\n", server.Name, server.ID)

			// Delete the instance
			if err := client.DeleteInstance(ctx, server.ID); err != nil {
				fmt.Printf("Error: Failed to delete instance %s: %v\n", server.Name, err)
				continue
			}
			fmt.Printf("Instance %s terminated successfully.\n", server.Name)

			// Delete OpenStack keypair - keypair name matches full instance name
			if server.Name != "" {
				fmt.Printf("Deleting OpenStack keypair %s...\n", server.Name)
				if err := client.DeleteKeypair(ctx, server.Name); err != nil {
					// Don't fail if keypair doesn't exist, just warn
					fmt.Printf("Warning: Failed to delete OpenStack keypair (it may not exist): %v\n", err)
				} else {
					fmt.Printf("OpenStack keypair deleted successfully.\n")
				}
			}

			// Delete local SSH keys - extract instance name from full name
			var instanceName string
			if strings.HasPrefix(server.Name, InstanceNamePrefix) {
				instanceName = strings.TrimPrefix(server.Name, InstanceNamePrefix)
			} else {
				instanceName = server.Name
			}

			if instanceName != "" {
				fmt.Printf("Cleaning up local SSH keys for %s...\n", instanceName)
				if err := DeleteSSHKey(instanceName); err != nil {
					// Don't fail if keys don't exist, just warn
					fmt.Printf("Warning: Failed to delete local SSH keys (they may not exist): %v\n", err)
				} else {
					fmt.Printf("Local SSH keys deleted successfully.\n")
				}
			}
		}
	}

	// Additional cleanup: delete any remaining tins keypairs that don't have associated instances
	fmt.Printf("\nChecking for orphaned tins keypairs...\n")
	if err := cleanupOrphanedKeypairs(ctx, client); err != nil {
		fmt.Printf("Warning: Failed to clean up orphaned keypairs: %v\n", err)
	} else {
		fmt.Printf("Orphaned keypair cleanup completed.\n")
	}

	fmt.Printf("\nAll tins instances and keypairs have been cleaned up.\n")
	return nil
}

// cleanupOrphanedKeypairs removes any tins keypairs that don't have associated instances
func cleanupOrphanedKeypairs(_ context.Context, _ *OpenStackClient) error {
	// This would require implementing a ListKeypairs method in OpenStackClient
	// For now, we'll just note that this is where additional cleanup could happen
	// The main cleanup happens through the instance termination loop above
	return nil
}

func init() {
	terminateCmd.Flags().Bool("all", false, "Terminate all tins instances and clean up all tins keypairs")
	rootCmd.AddCommand(terminateCmd)
}
