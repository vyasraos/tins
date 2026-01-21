package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

// extractIPAddress extracts the first available IP address from server addresses
func extractIPAddress(server *servers.Server) (string, error) {
	if len(server.Addresses) == 0 {
		return "", fmt.Errorf("no IP addresses found")
	}

	for _, addrList := range server.Addresses {
		// Type assert to []interface{} and then extract address info
		if addresses, ok := addrList.([]interface{}); ok {
			for _, addrInterface := range addresses {
				if addrMap, ok := addrInterface.(map[string]interface{}); ok {
					// Prefer fixed or floating IPs
					if addrType, ok := addrMap["OS-EXT-IPS:type"].(string); ok {
						if addrType == "fixed" || addrType == "floating" {
							if addr, ok := addrMap["addr"].(string); ok {
								return addr, nil
							}
						}
					} else if addr, ok := addrMap["addr"].(string); ok {
						// Fallback: use any address if type is not available
						return addr, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no valid IP address found in network %v", server.Addresses)
}

// showInstanceMenu displays instances using termbox-go and returns the selected instance
func showInstanceMenu(serverList []servers.Server) (*servers.Server, error) {
	if len(serverList) == 0 {
		return nil, fmt.Errorf("no instances available")
	}

	// Prepare display strings
	options := make([]string, len(serverList))
	for i := range serverList {
		ip := "N/A"
		if addr, err := extractIPAddress(&serverList[i]); err == nil {
			ip = addr
		}
		options[i] = fmt.Sprintf("%s | %s | %s", serverList[i].Name, serverList[i].Status, ip)
	}

	// Initialize termbox
	err := termbox.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize termbox: %w", err)
	}
	defer termbox.Close()

	selected := 0
	startY := 2 // Start below the prompt

	for {
		// Clear screen
		if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
			// If clearing fails, continue anyway as it's not critical
			continue
		}

		// Print prompt
		prompt := "? Select an instance (Use arrow keys)"
		for i, r := range prompt {
			termbox.SetCell(i, 0, r, termbox.ColorDefault, termbox.ColorDefault)
		}

		// Print options
		for i, option := range options {
			y := startY + i
			prefix := "  "
			if i == selected {
				prefix = "❯ "
			}

			// Print prefix
			for j, r := range prefix {
				termbox.SetCell(j, y, r, termbox.ColorDefault, termbox.ColorDefault)
			}

			// Print option text
			for j, r := range option {
				termbox.SetCell(len(prefix)+j, y, r, termbox.ColorDefault, termbox.ColorDefault)
			}
		}

		// Flush to screen
		termbox.Flush()

		// Handle input
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				if selected > 0 {
					selected--
				}
			case termbox.KeyArrowDown:
				if selected < len(options)-1 {
					selected++
				}
			case termbox.KeyEnter:
				return &serverList[selected], nil
			case termbox.KeyEsc:
				return nil, fmt.Errorf("cancelled by user")
			case termbox.KeyCtrlC:
				return nil, fmt.Errorf("cancelled by user")
			}
		case termbox.EventError:
			return nil, fmt.Errorf("termbox error: %v", ev.Err)
		}
	}
}

var connectCmd = &cobra.Command{
	Use:   "connect [instance-name-or-id]",
	Short: "Connect to a temporary instance via SSH",
	Long:  "Connect to a temporary instance via SSH. If no instance is specified, an interactive menu will be shown.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		var selectedServer *servers.Server

		if len(args) > 0 && args[0] != "" {
			// User provided instance identifier
			instanceIdentifier := args[0]

			// Try to get instance by ID first
			server, err := client.GetInstance(ctx, instanceIdentifier)
			if err == nil {
				// Check if it's a tins instance
				if server.Metadata != nil {
					if val, ok := server.Metadata[TempInstanceTag]; ok && val == "true" {
						selectedServer = server
					}
				}
				// Also check by name prefix
				if selectedServer == nil && len(server.Name) >= len(InstanceNamePrefix) && server.Name[:len(InstanceNamePrefix)] == InstanceNamePrefix {
					selectedServer = server
				}
			}

			// If not found by ID, search by name
			if selectedServer == nil {
				servers, err := client.ListInstances(ctx)
				if err != nil {
					return fmt.Errorf("failed to list instances: %w", err)
				}

				for i := range servers {
					if servers[i].Name == instanceIdentifier || servers[i].ID == instanceIdentifier {
						selectedServer = &servers[i]
						break
					}
					// Also check without prefix
					if strings.HasPrefix(servers[i].Name, InstanceNamePrefix) {
						nameWithoutPrefix := strings.TrimPrefix(servers[i].Name, InstanceNamePrefix)
						if nameWithoutPrefix == instanceIdentifier {
							selectedServer = &servers[i]
							break
						}
					}
				}
			}

			if selectedServer == nil {
				return fmt.Errorf("instance '%s' not found", instanceIdentifier)
			}
		} else {
			// Show interactive menu
			servers, err := client.ListInstances(ctx)
			if err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}

			if len(servers) == 0 {
				return fmt.Errorf("no instances available")
			}

			selectedServer, err = showInstanceMenu(servers)
			if err != nil {
				return err
			}
		}

		// Extract instance name from full name
		instanceName := strings.TrimPrefix(selectedServer.Name, InstanceNamePrefix)

		// Get IP address
		ipAddress, err := extractIPAddress(selectedServer)
		if err != nil {
			return fmt.Errorf("failed to get IP address for instance %s: %w", selectedServer.Name, err)
		}

		// Get SSH key path
		sshKeyPath := GetSSHKeyPath(instanceName)
		if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key not found at %s. The instance may have been created outside of this tool.", sshKeyPath)
		}

		// Build SSH command
		sshArgs := []string{
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"root@" + ipAddress,
		}

		// Add any additional SSH arguments from command line
		if len(args) > 1 {
			sshArgs = append(sshArgs, args[1:]...)
		}

		fmt.Printf("Connecting to %s (%s) using key %s...\n", selectedServer.Name, ipAddress, sshKeyPath)
		fmt.Println()

		// Execute SSH command
		sshCmd := exec.Command("ssh", sshArgs...)
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		if err := sshCmd.Run(); err != nil {
			return fmt.Errorf("SSH connection failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
