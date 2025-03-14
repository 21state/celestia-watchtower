package cmd

import (
	"fmt"
	"os"

	"github.com/21state/celestia-watchtower/config"
	"github.com/21state/celestia-watchtower/monitor"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start monitoring",
	Long:  `Start monitoring your Celestia node.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStart()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

// runStart starts the monitoring engine
func runStart() {
	// Load configuration
	fmt.Println("Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		fmt.Println("Please run 'celestia-watchtower setup' first.")
		os.Exit(1)
	}

	// Print configuration details
	fmt.Println("Configuration loaded successfully")
	fmt.Printf("RPC Endpoint: '%s'\n", cfg.Node.RPCEndpoint)
	fmt.Printf("Auth Token: '%s'\n", cfg.Node.AuthToken != "")
	fmt.Printf("Check Interval: %d seconds\n", cfg.Monitoring.CheckInterval)

	// Set log level based on config and debug flag
	isDebugMode := debugMode || cfg.Logging.Level == "debug"
	fmt.Printf("Debug Mode: %v\n", isDebugMode)

	// Create monitoring engine
	fmt.Println("Creating monitoring engine...")
	engine, err := monitor.NewEngine(cfg, isDebugMode)
	if err != nil {
		fmt.Printf("Error creating monitoring engine: %v\n", err)
		os.Exit(1)
	}

	// Start monitoring
	fmt.Println("Starting monitoring...")
	if err := engine.Start(); err != nil {
		fmt.Printf("Error starting monitoring: %v\n", err)
		os.Exit(1)
	}
}
