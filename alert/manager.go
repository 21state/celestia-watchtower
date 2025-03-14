package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/21state/celestia-watchtower/config"
)

// Manager handles sending alerts to configured channels
type Manager struct {
	config *config.Config
}

// NewManager creates a new alert manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// SendAlert sends an alert to all configured channels
func (m *Manager) SendAlert(message string) error {
	if !m.config.Alerts.Enabled {
		return nil
	}

	var errors []string

	// Send Telegram alert
	if m.config.Alerts.Telegram.Enabled {
		if err := m.sendTelegramAlert(message); err != nil {
			errors = append(errors, fmt.Sprintf("Telegram: %v", err))
		}
	}

	// Send Discord alert
	if m.config.Alerts.Discord.Enabled {
		if err := m.sendDiscordAlert(message); err != nil {
			errors = append(errors, fmt.Sprintf("Discord: %v", err))
		}
	}

	// Send Twilio SMS alert
	if m.config.Alerts.Twilio.Enabled {
		if err := m.sendTwilioAlert(message); err != nil {
			errors = append(errors, fmt.Sprintf("Twilio: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send alerts: %s", strings.Join(errors, "; "))
	}

	return nil
}

// sendTelegramAlert sends an alert via Telegram
func (m *Manager) sendTelegramAlert(message string) error {
	botToken := m.config.Alerts.Telegram.BotToken
	chatID := m.config.Alerts.Telegram.ChatID

	if botToken == "" || chatID == "" {
		return fmt.Errorf("Telegram bot token or chat ID not configured")
	}

	// Prepare API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	// Prepare request body
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", message)
	data.Set("parse_mode", "Markdown")

	// Send request
	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return fmt.Errorf("failed to send Telegram alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API returned non-OK status: %s", resp.Status)
	}

	return nil
}

// sendDiscordAlert sends an alert via Discord webhook
func (m *Manager) sendDiscordAlert(message string) error {
	webhook := m.config.Alerts.Discord.Webhook

	if webhook == "" {
		return fmt.Errorf("Discord webhook not configured")
	}

	// Prepare request body
	payload := map[string]interface{}{
		"content": message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	// Send request
	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send Discord alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Discord API returned non-OK status: %s", resp.Status)
	}

	return nil
}

// sendTwilioAlert sends an alert via Twilio SMS
func (m *Manager) sendTwilioAlert(message string) error {
	accountSID := m.config.Alerts.Twilio.AccountSID
	authToken := m.config.Alerts.Twilio.AuthToken
	fromNumber := m.config.Alerts.Twilio.FromNumber
	toNumber := m.config.Alerts.Twilio.ToNumber

	if accountSID == "" || authToken == "" || fromNumber == "" || toNumber == "" {
		return fmt.Errorf("Twilio credentials or phone numbers not configured")
	}

	// Prepare API URL
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	// Prepare request body
	data := url.Values{}
	data.Set("From", fromNumber)
	data.Set("To", toNumber)
	data.Set("Body", message)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Twilio request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Twilio alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Twilio API returned non-Created status: %s", resp.Status)
	}

	return nil
}

// TestAlert sends a test alert to verify alert configuration
func (m *Manager) TestAlert() error {
	message := "🔔 This is a test alert from Celestia Watchtower.\n\nIf you're receiving this, your alert configuration is working correctly!"
	return m.SendAlert(message)
}
