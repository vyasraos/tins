package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all temporary instances",
	Long:  "List all ephemeral OpenStack instances tagged as temporary instances.",
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

		// List instances
		servers, err := client.ListInstances(ctx)
		if err != nil {
			return fmt.Errorf("failed to list instances: %w", err)
		}

		if len(servers) == 0 {
			fmt.Println("No temporary instances found.")
			return nil
		}

		// Display instances in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED\t")
		fmt.Fprintln(w, "---\t----\t------\t-------\t")

		for _, server := range servers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n",
				server.ID,
				server.Name,
				server.Status,
				server.Created.Format("2006-01-02 15:04:05"),
			)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
