package monitor

import (
	"fmt"
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
