package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Node struct {
		RPCEndpoint string `yaml:"rpc_endpoint" mapstructure:"rpc_endpoint"`
		AuthToken   string `yaml:"auth_token" mapstructure:"auth_token"`
	} `yaml:"node" mapstructure:"node"`

	Monitoring struct {
		CheckInterval int `yaml:"check_interval" mapstructure:"check_interval"` // in seconds
	} `yaml:"monitoring" mapstructure:"monitoring"`

	Alerts struct {
		Enabled bool `yaml:"enabled" mapstructure:"enabled"`

		Telegram struct {
			Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
			BotToken string `yaml:"bot_token" mapstructure:"bot_token"`
			ChatID   string `yaml:"chat_id" mapstructure:"chat_id"`
		} `yaml:"telegram" mapstructure:"telegram"`

		Discord struct {
			Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
			Webhook string `yaml:"webhook" mapstructure:"webhook"`
		} `yaml:"discord" mapstructure:"discord"`

		Twilio struct {
			Enabled     bool   `yaml:"enabled" mapstructure:"enabled"`
			AccountSID  string `yaml:"account_sid" mapstructure:"account_sid"`
			AuthToken   string `yaml:"auth_token" mapstructure:"auth_token"`
			FromNumber  string `yaml:"from_number" mapstructure:"from_number"`
			ToNumber    string `yaml:"to_number" mapstructure:"to_number"`
		} `yaml:"twilio" mapstructure:"twilio"`
	} `yaml:"alerts" mapstructure:"alerts"`

	Thresholds struct {
		SyncStatus struct {
			BlocksBehindCritical int `yaml:"blocks_behind_critical" mapstructure:"blocks_behind_critical"`
		} `yaml:"sync_status" mapstructure:"sync_status"`

		Network struct {
			MinPeersHealthy int `yaml:"min_peers_healthy" mapstructure:"min_peers_healthy"`
		} `yaml:"network" mapstructure:"network"`
	} `yaml:"thresholds" mapstructure:"thresholds"`

	Logging struct {
		Level string `yaml:"level" mapstructure:"level"` // info, debug
	} `yaml:"logging" mapstructure:"logging"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	cfg := &Config{}

	// Node defaults
	cfg.Node.RPCEndpoint = "http://localhost:26658"
	cfg.Node.AuthToken = ""

	// Monitoring defaults
	cfg.Monitoring.CheckInterval = 60 // 1 minute

	// Alerts defaults
	cfg.Alerts.Enabled = false
	cfg.Alerts.Telegram.Enabled = false
	cfg.Alerts.Telegram.BotToken = ""
	cfg.Alerts.Telegram.ChatID = ""
	
	// Discord alerts
	cfg.Alerts.Discord.Enabled = false
	cfg.Alerts.Discord.Webhook = ""
	
	// Twilio alerts
	cfg.Alerts.Twilio.Enabled = false
	cfg.Alerts.Twilio.AccountSID = ""
	cfg.Alerts.Twilio.AuthToken = ""
	cfg.Alerts.Twilio.FromNumber = ""
	cfg.Alerts.Twilio.ToNumber = ""

	// Threshold defaults
	cfg.Thresholds.SyncStatus.BlocksBehindCritical = 10
	cfg.Thresholds.Network.MinPeersHealthy = 5

	// Logging defaults
	cfg.Logging.Level = "info"

	return cfg
}

// ConfigDir returns the path to the configuration directory
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".celestia-watchtower"), nil
}

// ConfigFile returns the path to the configuration file
func ConfigFile() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// TempStatusFile returns the path to the temporary status file
func TempStatusFile() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "status.json"), nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(cfg *Config) error {
	configFile, err := ConfigFile()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configFile, err := ConfigFile()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found, run 'celestia-watchtower setup' first")
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}
