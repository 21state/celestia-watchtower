package rpc

import (
	"context"
	"fmt"

	openrpc "github.com/celestiaorg/celestia-openrpc"
)

// Client is a wrapper around the celestia-openrpc client
type Client struct {
	client *openrpc.Client
	ctx    context.Context
}

// BandwidthStats represents bandwidth statistics
type BandwidthStats struct {
	TotalIn  int64   // Total bytes in
	TotalOut int64   // Total bytes out
	RateIn   float64 // Bytes in per second
	RateOut  float64 // Bytes out per second
}

// NewClient creates a new RPC client
func NewClient(ctx context.Context, rpcEndpoint, authToken string) (*Client, error) {
	// Validate the RPC endpoint
	if rpcEndpoint == "" {
		return nil, fmt.Errorf("RPC endpoint cannot be empty")
	}

	client, err := openrpc.NewClient(ctx, rpcEndpoint, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetNetworkHead returns the network head height
func (c *Client) GetNetworkHead() (uint64, error) {
	header, err := c.client.Header.NetworkHead(c.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get network head: %w", err)
	}

	return header.Height(), nil
}

// GetLocalHead returns the local head height
func (c *Client) GetLocalHead() (uint64, error) {
	header, err := c.client.Header.LocalHead(c.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get local head: %w", err)
	}

	return header.Height(), nil
}

// GetSyncState returns the sync state as a string
func (c *Client) GetSyncState() (string, error) {
	syncState, err := c.client.Header.SyncState(c.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get sync state: %w", err)
	}

	// Convert the sync state to a string
	// Since we can't directly switch on the sync.State type,
	// we'll use the String() method if available, or return a generic description
	return fmt.Sprintf("%v", syncState), nil
}

// GetPeers returns the number of connected peers
func (c *Client) GetPeers() (int, error) {
	peers, err := c.client.P2P.Peers(c.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get peers: %w", err)
	}

	return len(peers), nil
}

// GetNATStatus returns the NAT status as a string
func (c *Client) GetNATStatus() (string, error) {
	natStatus, err := c.client.P2P.NATStatus(c.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get NAT status: %w", err)
	}

	return natStatus.String(), nil
}

// GetBandwidthStats returns bandwidth statistics
func (c *Client) GetBandwidthStats() (*BandwidthStats, error) {
	stats, err := c.client.P2P.BandwidthStats(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bandwidth stats: %w", err)
	}

	// The API returns the values directly, no need to parse
	return &BandwidthStats{
		TotalIn:  stats.TotalIn,
		TotalOut: stats.TotalOut,
		RateIn:   stats.RateIn,
		RateOut:  stats.RateOut,
	}, nil
}

// Close closes the client connection
func (c *Client) Close() {
	c.client.Close()
}
