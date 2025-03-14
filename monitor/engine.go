package monitor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/21state/celestia-watchtower/alert"
	"github.com/21state/celestia-watchtower/config"
	"github.com/21state/celestia-watchtower/rpc"
)

// Engine is responsible for monitoring the node
type Engine struct {
	client      *rpc.Client
	config      *config.Config
	alerter     *alert.Manager
	ctx         context.Context
	cancel      context.CancelFunc
	lastStatus  *Status
	isDebugMode bool
}

// NewEngine creates a new monitoring engine
func NewEngine(cfg *config.Config, isDebugMode bool) (*Engine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Validate configuration
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	// Debug output
	if isDebugMode {
		fmt.Printf("Debug: RPC Endpoint = '%s'\n", cfg.Node.RPCEndpoint)
		fmt.Printf("Debug: Auth Token = '%s'\n", cfg.Node.AuthToken)
	}

	// Validate RPC endpoint
	if cfg.Node.RPCEndpoint == "" {
		return nil, fmt.Errorf("RPC endpoint cannot be empty")
	}

	client, err := rpc.NewClient(ctx, cfg.Node.RPCEndpoint, cfg.Node.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	alerter := alert.NewManager(cfg)

	return &Engine{
		client:      client,
		config:      cfg,
		alerter:     alerter,
		ctx:         ctx,
		cancel:      cancel,
		isDebugMode: isDebugMode,
	}, nil
}

// Start starts the monitoring engine
func (e *Engine) Start() error {
	fmt.Println("🔭 Celestia Watchtower started - Developed by 21state")
	fmt.Printf("Monitoring %s every %d seconds\n", e.config.Node.RPCEndpoint, e.config.Monitoring.CheckInterval)

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for periodic checks
	ticker := time.NewTicker(time.Duration(e.config.Monitoring.CheckInterval) * time.Second)
	defer ticker.Stop()

	// Run initial check immediately
	if err := e.runCheck(); err != nil {
		logError("Initial check failed: %v", err)
	}

	// Main loop
	for {
		select {
		case <-ticker.C:
			if err := e.runCheck(); err != nil {
				logError("Check failed: %v", err)
			}
		case <-sigCh:
			fmt.Println("\nShutting down...")
			e.cancel()
			return nil
		case <-e.ctx.Done():
			return nil
		}
	}
}

// Stop stops the monitoring engine
func (e *Engine) Stop() {
	e.cancel()
}

// GetLastStatus returns the last known status
func (e *Engine) GetLastStatus() *Status {
	return e.lastStatus
}

// runCheck performs a single check of the node status
func (e *Engine) runCheck() error {
	// Check node status
	status, err := CheckNodeStatus(e.client, e.config)
	if err != nil {
		return err
	}

	// Save status
	if err := SaveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	// Check if we need to send alerts
	if e.lastStatus != nil && e.lastStatus.Healthy && !status.Healthy {
		// Node has become unhealthy, send alert
		if err := e.sendAlerts(status); err != nil {
			logError("Failed to send alerts: %v", err)
		}
	}

	// Print status
	e.printStatus(status)

	// Update last status
	e.lastStatus = status

	return nil
}

// printStatus prints the current node status
func (e *Engine) printStatus(status *Status) {
	timestamp := status.Timestamp.Format("2006-01-02 15:04:05")
	
	// Health indicator
	healthStatus := "✅ HEALTHY"
	if !status.Healthy {
		healthStatus = "❌ UNHEALTHY"
	}
	
	fmt.Printf("[%s] Status: %s\n", timestamp, healthStatus)
	
	// Sync status
	syncHealth := "✅"
	if !status.SyncHealthy {
		syncHealth = "❌"
	}
	fmt.Printf("  Sync: %s Height: %d/%d (diff: %d) State: %s\n", 
		syncHealth, status.LocalHeight, status.NetworkHeight, status.HeightDiff, status.SyncState)
	
	// Network status
	netHealth := "✅"
	if !status.NetHealthy {
		netHealth = "❌"
	}
	fmt.Printf("  Network: %s Peers: %d NAT: %s\n", 
		netHealth, status.PeerCount, status.NATStatus)
	
	// Bandwidth stats
	fmt.Printf("  Bandwidth: In: %.2f KB/s (Total: %d MB) Out: %.2f KB/s (Total: %d MB)\n",
		status.Bandwidth.RateIn/1024, status.Bandwidth.TotalIn/(1024*1024),
		status.Bandwidth.RateOut/1024, status.Bandwidth.TotalOut/(1024*1024))
	
	// Print debug info if in debug mode
	if e.isDebugMode {
		fmt.Println("  Debug info:")
		fmt.Printf("    SyncHealthy: %v (threshold: %d blocks)\n", 
			status.SyncHealthy, e.config.Thresholds.SyncStatus.BlocksBehindCritical)
		fmt.Printf("    NetHealthy: %v (threshold: %d peers)\n", 
			status.NetHealthy, e.config.Thresholds.Network.MinPeersHealthy)
	}
}

// sendAlerts sends alerts to all configured channels
func (e *Engine) sendAlerts(status *Status) error {
	var alertMessage string
	
	// Determine which check failed
	if !status.SyncHealthy {
		alertMessage = fmt.Sprintf("⚠️ Sync issue detected! Node is %d blocks behind the network.", status.HeightDiff)
	} else if !status.NetHealthy {
		alertMessage = fmt.Sprintf("⚠️ Network issue detected! Node has only %d peers (minimum: %d).", 
			status.PeerCount, e.config.Thresholds.Network.MinPeersHealthy)
	} else {
		alertMessage = "⚠️ Node is unhealthy! Check the logs for more details."
	}
	
	// Add timestamp and node info
	fullMessage := fmt.Sprintf("%s\nTimestamp: %s\nNode: %s\nLocal Height: %d\nNetwork Height: %d\nPeers: %d",
		alertMessage, status.Timestamp.Format(time.RFC1123),
		e.config.Node.RPCEndpoint, status.LocalHeight, status.NetworkHeight, status.PeerCount)
	
	// Log the alert
	fmt.Printf("Sending alert: %s\n", alertMessage)
	
	// Send alert
	return e.alerter.SendAlert(fullMessage)
}

// logError logs an error message
func logError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
}
