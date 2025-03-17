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
		return nil, fmt.Errorf("[ERROR] configuration is nil")
	}

	// Debug output
	if isDebugMode {
		fmt.Printf("[DEBUG] RPC Endpoint = '%s'\n", cfg.Node.RPCEndpoint)
		fmt.Printf("[DEBUG] Auth Token = '%s'\n", cfg.Node.AuthToken != "")
	}

	// Validate RPC endpoint
	if cfg.Node.RPCEndpoint == "" {
		return nil, fmt.Errorf("[ERROR] RPC endpoint cannot be empty")
	}

	client, err := rpc.NewClient(ctx, cfg.Node.RPCEndpoint, cfg.Node.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to create RPC client: %w", err)
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
	fmt.Println("[INFO] 🔭 Celestia Watchtower started")
	fmt.Printf("[INFO] Monitoring %s every %d seconds\n", e.config.Node.RPCEndpoint, e.config.Monitoring.CheckInterval)

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for periodic checks
	ticker := time.NewTicker(time.Duration(e.config.Monitoring.CheckInterval) * time.Second)
	defer ticker.Stop()

	// Initial check
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
			fmt.Println("[INFO] Shutting down...")
			e.Stop()
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
		return fmt.Errorf("[ERROR] failed to check node status: %w", err)
	}

	// Update last status
	e.lastStatus = status

	// Always print basic status in info mode
	e.printInfoStatus(status)
	
	// Print detailed status if debug mode is enabled
	if e.isDebugMode {
		e.printDebugStatus(status)
	}

	// Send alerts if needed
	if !status.Healthy && e.config.Alerts.Enabled {
		if err := e.sendAlerts(status); err != nil {
			return fmt.Errorf("[ERROR] failed to send alerts: %w", err)
		}
	}

	return nil
}

// formatDataSize formats a byte value into the most appropriate unit
// Returns the converted value and the unit string
func formatDataSize(bytes float64) (float64, string) {
    const (
        KB = 1024.0
        MB = KB * 1024.0
        GB = MB * 1024.0
        TB = GB * 1024.0
    )
    
    units := []struct {
        divisor float64
        unit    string
    }{
        {TB, "TB"},
        {GB, "GB"},
        {MB, "MB"},
        {KB, "KB"},
    }
    
    for _, u := range units {
        if bytes >= u.divisor {
            return bytes / u.divisor, u.unit
        }
    }
    
    return bytes, "B"
}

// formatBandwidth formats bandwidth metrics consistently
func formatBandwidth(status *Status) (inRate, outRate float64, inTotal, inUnit, outTotal, outUnit string) {
    // Convert rates to KB/s
    inRate = status.Bandwidth.RateIn / 1024.0
    outRate = status.Bandwidth.RateOut / 1024.0
    
    // Format totals with appropriate units
    inTotalVal, inUnit := formatDataSize(float64(status.Bandwidth.TotalIn))
    outTotalVal, outUnit := formatDataSize(float64(status.Bandwidth.TotalOut))
    
    return inRate, outRate, fmt.Sprintf("%.2f", inTotalVal), inUnit, fmt.Sprintf("%.2f", outTotalVal), outUnit
}

// printInfoStatus prints basic status information in info mode
func (e *Engine) printInfoStatus(status *Status) {
	timestamp := status.Timestamp.Format("2006-01-02 15:04:05")
	
	// Health indicator
	healthStatus := "[OK] HEALTHY"
	if !status.Healthy {
		healthStatus = "[!!] UNHEALTHY"
	}
	
	inRate, outRate, inTotal, inUnit, outTotal, outUnit := formatBandwidth(status)
	
	fmt.Printf("[INFO] [%s] Status: %s | Height: %d/%d | Peers: %d | NAT: %s | In: %.1f KB/s (%s %s) | Out: %.1f KB/s (%s %s)\n", 
		timestamp, 
		healthStatus, 
		status.LocalHeight, 
		status.NetworkHeight, 
		status.PeerCount,
		status.NATStatus,
		inRate, inTotal, inUnit,
		outRate, outTotal, outUnit)
}

// printDebugStatus prints detailed status information in debug mode
func (e *Engine) printDebugStatus(status *Status) {
	// Sync status
	syncHealth := "[OK]"
	if !status.SyncHealthy {
		syncHealth = "[!!]"
	}
	fmt.Printf("[DEBUG]   Sync: %s Height: %d/%d (diff: %d)\n", 
		syncHealth, status.LocalHeight, status.NetworkHeight, status.HeightDiff)
	
	// Network status
	netHealth := "[OK]"
	if !status.NetHealthy {
		netHealth = "[!!]"
	}
	fmt.Printf("[DEBUG]   Network: %s Peers: %d NAT: %s\n", 
		netHealth, status.PeerCount, status.NATStatus)
	
	inRate, outRate, inTotal, inUnit, outTotal, outUnit := formatBandwidth(status)
	
	fmt.Printf("[DEBUG]   Bandwidth: In: %.2f KB/s (Total: %s %s) Out: %.2f KB/s (Total: %s %s)\n",
		inRate, inTotal, inUnit,
		outRate, outTotal, outUnit)
}

// sendAlerts sends alerts to all configured channels
func (e *Engine) sendAlerts(status *Status) error {
	// Prepare alert message
	message := fmt.Sprintf("⚠️ Celestia Node Alert ⚠️\n\n")
	
	// Add timestamp
	message += fmt.Sprintf("Time: %s\n\n", status.Timestamp.Format("2006-01-02 15:04:05"))
	
	// Add sync status if unhealthy
	if !status.SyncHealthy {
		message += fmt.Sprintf("❌ Sync Issue: Node is %d blocks behind the network\n", status.HeightDiff)
		message += fmt.Sprintf("   Local Height: %d, Network Height: %d\n\n", status.LocalHeight, status.NetworkHeight)
	}
	
	// Add network status if unhealthy
	if !status.NetHealthy {
		message += fmt.Sprintf("❌ Network Issue: Node has only %d peers (min: %d)\n", 
			status.PeerCount, e.config.Thresholds.Network.MinPeersHealthy)
		message += fmt.Sprintf("   NAT Status: %s\n\n", status.NATStatus)
	}
	
	// Send alert
	if err := e.alerter.SendAlert(message); err != nil {
		return fmt.Errorf("[ERROR] failed to send alert: %w", err)
	}
	
	return nil
}

// logError logs an error message
func logError(format string, args ...interface{}) {
	fmt.Printf("[ERROR] %s\n", fmt.Sprintf(format, args...))
}
