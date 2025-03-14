package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/21state/celestia-watchtower/config"
	"github.com/21state/celestia-watchtower/monitor"
	"github.com/spf13/cobra"
)

var watchFlag bool

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show node status",
	Long:  `Show the current status of your Celestia node.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "Watch for status updates")
}

// runStatus shows the current node status
func runStatus() {
	// Check if status file exists
	status, err := monitor.LoadStatus()
	if err != nil {
		fmt.Printf("[ERROR] Error loading status: %v\n", err)
		fmt.Println("[INFO] Please run 'celestia-watchtower start' first.")
		os.Exit(1)
	}

	// Print status
	printStatus(status)

	// Check if we should watch for updates
	if !watchFlag {
		return
	}

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("[INFO] Watching for status updates. Press Ctrl+C to exit.")

	// Load config to get check interval
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("[ERROR] Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create ticker for periodic checks
	ticker := time.NewTicker(time.Duration(cfg.Monitoring.CheckInterval) * time.Second)
	defer ticker.Stop()

	// Main loop
	for {
		select {
		case <-ticker.C:
			// Load updated status
			newStatus, err := monitor.LoadStatus()
			if err != nil {
				fmt.Printf("[ERROR] Error loading status: %v\n", err)
				continue
			}

			// Clear screen and print updated status
			fmt.Print("\033[H\033[2J") // ANSI escape sequence to clear screen
			printStatus(newStatus)
			fmt.Println("[INFO] Watching for status updates. Press Ctrl+C to exit.")
		case <-sigCh:
			fmt.Println("[INFO] Exiting...")
			return
		}
	}
}

// printStatus prints the current node status
func printStatus(status *monitor.Status) {
	timestamp := status.Timestamp.Format("2006-01-02 15:04:05")
	
	// Health indicator
	healthStatus := "[OK] HEALTHY"
	if !status.Healthy {
		healthStatus = "[!!] UNHEALTHY"
	}
	
	fmt.Printf("[INFO] [%s] Status: %s\n", timestamp, healthStatus)
	
	// Sync status
	syncHealth := "[OK]"
	if !status.SyncHealthy {
		syncHealth = "[!!]"
	}
	fmt.Printf("[INFO]   Sync: %s Height: %d/%d (diff: %d)\n", 
		syncHealth, status.LocalHeight, status.NetworkHeight, status.HeightDiff)
	
	// Network status
	netHealth := "[OK]"
	if !status.NetHealthy {
		netHealth = "[!!]"
	}
	fmt.Printf("[INFO]   Network: %s Peers: %d NAT: %s\n", 
		netHealth, status.PeerCount, status.NATStatus)
	
	// Bandwidth stats
	fmt.Printf("[INFO]   Bandwidth: In: %.2f KB/s (Total: %d MB) Out: %.2f KB/s (Total: %d MB)\n",
		status.Bandwidth.RateIn/1024, status.Bandwidth.TotalIn/(1024*1024),
		status.Bandwidth.RateOut/1024, status.Bandwidth.TotalOut/(1024*1024))
}
