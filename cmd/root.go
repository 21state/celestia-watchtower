package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags
	debugMode bool

	// rootCmd represents the base command
	rootCmd = &cobra.Command{
		Use:   "celestia-watchtower",
		Short: "Monitor your Celestia node",
		Long: `Celestia Watchtower is a monitoring tool for Celestia nodes.
It checks the node's status periodically and sends alerts if issues are detected.`,
	}
)

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode")
}
