package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build via ldflags
	Version = "dev"
	// Commit is set during build via ldflags
	Commit = "unknown"
)

const (
	// InstanceNamePrefix is the prefix used for all temporary instances
	InstanceNamePrefix = "tins-"
	// TempInstanceTag is the metadata tag used to identify temporary instances
	TempInstanceTag = "tins"
)

var rootCmd = &cobra.Command{
	Use:   "tins",
	Short: "A CLI tool for managing temporary OpenStack instances",
	Long:  "tins (Temporary Instances) creates and manages ephemeral OpenStack instances with automatic SSH key management.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  "Print the version number and build information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tins version %s (commit: %s)\n", Version, Commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
