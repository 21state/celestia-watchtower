package cmd

import (
	"fmt"
	"os"

	"github.com/21state/celestia-watchtower/alert"
	"github.com/21state/celestia-watchtower/config"
	"github.com/spf13/cobra"
)

// testAlertCmd represents the test-alert command
var testAlertCmd = &cobra.Command{
	Use:   "test-alert",
	Short: "Test alert notifications",
	Long:  `Send a test alert to verify that alert notifications are working correctly.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTestAlert()
	},
}

func init() {
	rootCmd.AddCommand(testAlertCmd)
}

// runTestAlert sends a test alert
func runTestAlert() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		fmt.Println("Please run 'celestia-watchtower setup' first.")
		os.Exit(1)
	}

	// Check if alerts are enabled
	if !cfg.Alerts.Enabled {
		fmt.Println("Alerts are disabled in the configuration.")
		fmt.Println("Please enable alerts with 'celestia-watchtower setup' first.")
		os.Exit(1)
	}

	// Check if at least one alert channel is configured
	if !cfg.Alerts.Telegram.Enabled && !cfg.Alerts.Discord.Enabled && !cfg.Alerts.Twilio.Enabled {
		fmt.Println("No alert channels are enabled in the configuration.")
		fmt.Println("Please configure at least one alert channel with 'celestia-watchtower setup'.")
		os.Exit(1)
	}

	fmt.Println("Sending test alert...")

	// Create alert manager and send test alert
	alerter := alert.NewManager(cfg)
	if err := alerter.TestAlert(); err != nil {
		fmt.Printf("Error sending test alert: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Test alert sent successfully!")
	fmt.Println("If you don't receive the alert, please check your configuration.")
}
