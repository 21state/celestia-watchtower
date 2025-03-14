package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/21state/celestia-watchtower/config"
	"github.com/spf13/cobra"
)

// reloadCmd represents the reload command
var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the watchtower configuration",
	Long:  `Reload the watchtower configuration and restart the service if running as a systemd service.`,
	Run: func(cmd *cobra.Command, args []string) {
		runReload()
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}

// runReload reloads the configuration
func runReload() {
	// Check if config file exists
	configFile, err := config.ConfigFile()
	if err != nil {
		fmt.Printf("Error getting config file path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Configuration file not found.")
		fmt.Println("Please run 'celestia-watchtower setup' first.")
		os.Exit(1)
	}

	// Try to reload the config to validate it
	_, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration loaded successfully.")

	// Check if running as a systemd service
	if isRunningAsSystemd() {
		fmt.Println("Detected running as a systemd service. Restarting service...")
		
		// Restart the service
		cmd := exec.Command("systemctl", "restart", "celestia-watchtower.service")
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error restarting service: %v\n", err)
			fmt.Println("You may need to restart the service manually.")
			os.Exit(1)
		}
		
		fmt.Println("Service restarted successfully.")
	} else {
		fmt.Println("Not running as a systemd service.")
		fmt.Println("Please restart the watchtower manually with 'celestia-watchtower start'.")
	}
}

// isRunningAsSystemd checks if the process is running as a systemd service
func isRunningAsSystemd() bool {
	// This only works on Linux
	if runtime.GOOS != "linux" {
		return false
	}

	// Check if the process is running under systemd
	ppid := os.Getppid()
	ppidCmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", ppid))
	if err != nil {
		return false
	}

	cmdline := string(ppidCmdline)
	return strings.Contains(cmdline, "systemd")
}

// findServiceFile tries to find the systemd service file
func findServiceFile() (string, error) {
	// Common locations for systemd service files
	locations := []string{
		"/etc/systemd/system/celestia-watchtower.service",
		"/lib/systemd/system/celestia-watchtower.service",
		"/usr/lib/systemd/system/celestia-watchtower.service",
	}

	// Check if the service file exists in any of the locations
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location, nil
		}
	}

	// Check in user's systemd directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	userServiceDir := filepath.Join(homeDir, ".config/systemd/user")
	userServiceFile := filepath.Join(userServiceDir, "celestia-watchtower.service")
	
	if _, err := os.Stat(userServiceFile); err == nil {
		return userServiceFile, nil
	}

	return "", fmt.Errorf("service file not found")
}
