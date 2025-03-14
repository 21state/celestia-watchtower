package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/21state/celestia-watchtower/config"
	"github.com/21state/celestia-watchtower/rpc"
)

// Status represents the node status
type Status struct {
	Timestamp time.Time `json:"timestamp"`
	
	// Sync status
	NetworkHeight uint64 `json:"network_height"`
	LocalHeight   uint64 `json:"local_height"`
	HeightDiff    int64  `json:"height_diff"`
	SyncHealthy   bool   `json:"sync_healthy"`
	
	// Network status
	PeerCount   int    `json:"peer_count"`
	NATStatus   string `json:"nat_status"`
	NetHealthy  bool   `json:"net_healthy"`
	
	// Bandwidth stats
	Bandwidth struct {
		TotalIn  int64   `json:"total_in"`
		TotalOut int64   `json:"total_out"`
		RateIn   float64 `json:"rate_in"`
		RateOut  float64 `json:"rate_out"`
	} `json:"bandwidth"`
	
	// Overall status
	Healthy bool `json:"healthy"`
}

// SaveStatus saves the current status to a temporary file
func SaveStatus(status *Status) error {
	statusFile, err := config.TempStatusFile()
	if err != nil {
		return fmt.Errorf("[ERROR] failed to get status file path: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir, err := config.ConfigDir()
	if err != nil {
		return fmt.Errorf("[ERROR] failed to get config directory: %w", err)
	}
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("[ERROR] failed to create config directory: %w", err)
	}

	// Marshal status to JSON
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to marshal status: %w", err)
	}

	// Write status file
	if err := os.WriteFile(statusFile, data, 0644); err != nil {
		return fmt.Errorf("[ERROR] failed to write status file: %w", err)
	}

	return nil
}

// LoadStatus loads the current status from the temporary file
func LoadStatus() (*Status, error) {
	statusFile, err := config.TempStatusFile()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to get status file path: %w", err)
	}

	// Check if status file exists
	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("[ERROR] status file not found, run 'celestia-watchtower start' first")
	}

	// Read status file
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to read status file: %w", err)
	}

	// Parse status
	status := &Status{}
	if err := json.Unmarshal(data, status); err != nil {
		return nil, fmt.Errorf("[ERROR] failed to parse status file: %w", err)
	}

	return status, nil
}

// CheckNodeStatus checks the node status and returns a Status object
func CheckNodeStatus(client *rpc.Client, cfg *config.Config) (*Status, error) {
	status := &Status{
		Timestamp: time.Now(),
	}

	// Check network height
	networkHeight, err := client.GetNetworkHead()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to get network height: %w", err)
	}
	status.NetworkHeight = networkHeight
	
	// Check local height
	localHeight, err := client.GetLocalHead()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to get local height: %w", err)
	}
	status.LocalHeight = localHeight
	
	// Calculate height difference
	status.HeightDiff = int64(networkHeight) - int64(localHeight)
	
	// Check sync health
	status.SyncHealthy = status.HeightDiff <= int64(cfg.Thresholds.SyncStatus.BlocksBehindCritical)
	
	// Check peer count
	peerCount, err := client.GetPeers()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to get peer count: %w", err)
	}
	status.PeerCount = peerCount
	
	// Check NAT status
	natStatus, err := client.GetNATStatus()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to get NAT status: %w", err)
	}
	status.NATStatus = natStatus
	
	// Check network health
	status.NetHealthy = peerCount >= cfg.Thresholds.Network.MinPeersHealthy
	
	// Check bandwidth stats
	bandwidthStats, err := client.GetBandwidthStats()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to get bandwidth stats: %w", err)
	}
	status.Bandwidth.TotalIn = bandwidthStats.TotalIn
	status.Bandwidth.TotalOut = bandwidthStats.TotalOut
	status.Bandwidth.RateIn = bandwidthStats.RateIn
	status.Bandwidth.RateOut = bandwidthStats.RateOut
	
	// Overall health
	status.Healthy = status.SyncHealthy && status.NetHealthy
	
	return status, nil
}
