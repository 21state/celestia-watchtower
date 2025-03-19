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
	fmt.Println("[INFO] Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("[ERROR] Error loading configuration: %v\n", err)
		fmt.Println("[INFO] Please run 'celestia-watchtower setup' first.")
		os.Exit(1)
	}

	// Print configuration details
	fmt.Println("[INFO] Configuration loaded successfully")
	fmt.Printf("[INFO] RPC Endpoint: '%s'\n", cfg.Node.RPCEndpoint)
	fmt.Printf("[INFO] Auth Token: %v\n", cfg.Node.AuthToken != "")
	fmt.Printf("[INFO] Check Interval: %d seconds\n", cfg.Monitoring.CheckInterval)

	// Create monitoring engine
	fmt.Println("[INFO] Creating monitoring engine...")
	engine, err := monitor.NewEngine(cfg)
	if err != nil {
		fmt.Printf("[ERROR] Error creating monitoring engine: %v\n", err)
		os.Exit(1)
	}

	// Start monitoring
	fmt.Println("[INFO] Starting monitoring engine...")
	if err := engine.Start(); err != nil {
		fmt.Printf("[ERROR] Error starting monitoring engine: %v\n", err)
		os.Exit(1)
	}
}
