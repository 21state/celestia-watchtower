package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/21state/celestia-watchtower/config"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up the watchtower configuration",
	Long:  `Set up the watchtower configuration interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSetup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// runSetup runs the setup process
func runSetup() {
	fmt.Println("🔧 Celestia Watchtower Setup")
	fmt.Println("This wizard will help you configure the watchtower.")
	fmt.Println("Press Enter to accept the default values shown in [brackets].")
	fmt.Println()

	// Load default config
	cfg := config.DefaultConfig()

	// Create a reader for user input
	reader := bufio.NewReader(os.Stdin)

	// Node settings
	fmt.Println("📡 Node Settings")
	cfg.Node.RPCEndpoint = promptString(reader, "RPC Endpoint", cfg.Node.RPCEndpoint)
	cfg.Node.AuthToken = promptString(reader, "Auth Token", cfg.Node.AuthToken)
	fmt.Println()

	// Monitoring settings
	fmt.Println("⏱️ Monitoring Settings")
	checkInterval := promptInt(reader, "Check Interval (seconds)", cfg.Monitoring.CheckInterval)
	cfg.Monitoring.CheckInterval = checkInterval
	fmt.Println()

	// Threshold settings
	fmt.Println("🎚️ Threshold Settings")
	blocksBehind := promptInt(reader, "Critical Blocks Behind", cfg.Thresholds.SyncStatus.BlocksBehindCritical)
	cfg.Thresholds.SyncStatus.BlocksBehindCritical = blocksBehind

	minPeers := promptInt(reader, "Minimum Healthy Peers", cfg.Thresholds.Network.MinPeersHealthy)
	cfg.Thresholds.Network.MinPeersHealthy = minPeers
	fmt.Println()

	// Alert settings
	fmt.Println("🔔 Alert Settings")
	enableAlerts := promptBool(reader, "Enable Alerts", cfg.Alerts.Enabled)
	cfg.Alerts.Enabled = enableAlerts

	if enableAlerts {
		// Telegram alerts
		enableTelegram := promptBool(reader, "Enable Telegram Alerts", cfg.Alerts.Telegram.Enabled)
		cfg.Alerts.Telegram.Enabled = enableTelegram

		if enableTelegram {
			cfg.Alerts.Telegram.BotToken = promptString(reader, "Telegram Bot Token", cfg.Alerts.Telegram.BotToken)
			cfg.Alerts.Telegram.ChatID = promptString(reader, "Telegram Chat ID", cfg.Alerts.Telegram.ChatID)
		}

		// Discord alerts
		enableDiscord := promptBool(reader, "Enable Discord Alerts", cfg.Alerts.Discord.Enabled)
		cfg.Alerts.Discord.Enabled = enableDiscord

		if enableDiscord {
			cfg.Alerts.Discord.Webhook = promptString(reader, "Discord Webhook URL", cfg.Alerts.Discord.Webhook)
		}

		// Twilio alerts
		enableTwilio := promptBool(reader, "Enable SMS Alerts (Twilio)", cfg.Alerts.Twilio.Enabled)
		cfg.Alerts.Twilio.Enabled = enableTwilio

		if enableTwilio {
			cfg.Alerts.Twilio.AccountSID = promptString(reader, "Twilio Account SID", cfg.Alerts.Twilio.AccountSID)
			cfg.Alerts.Twilio.AuthToken = promptString(reader, "Twilio Auth Token", cfg.Alerts.Twilio.AuthToken)
			cfg.Alerts.Twilio.FromNumber = promptString(reader, "Twilio From Number", cfg.Alerts.Twilio.FromNumber)
			cfg.Alerts.Twilio.ToNumber = promptString(reader, "Twilio To Number", cfg.Alerts.Twilio.ToNumber)
		}
	}
	fmt.Println()

	// Logging settings
	fmt.Println("📝 Logging Settings")
	fmt.Println("Use the 'celestia-watchtower start --debug' flag to enable debug logging.")
	fmt.Println()

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	configFile, _ := config.ConfigFile()
	fmt.Printf("✅ Configuration saved to %s\n", configFile)
	fmt.Println("You can now start the watchtower with 'celestia-watchtower start'")
}

// promptString prompts the user for a string value
func promptString(reader *bufio.Reader, prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// promptInt prompts the user for an integer value
func promptInt(reader *bufio.Reader, prompt string, defaultValue int) int {
	fmt.Printf("%s [%d]: ", prompt, defaultValue)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("Invalid number, using default: %v\n", err)
		return defaultValue
	}
	return value
}

// promptBool prompts the user for a boolean value
func promptBool(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "no"
	if defaultValue {
		defaultStr = "yes"
	}

	fmt.Printf("%s (yes/no) [%s]: ", prompt, defaultStr)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return defaultValue
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultValue
	}

	return input == "yes" || input == "y" || input == "true"
}
