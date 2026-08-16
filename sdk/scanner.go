package sdk

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/network"
	"github.com/shahradelahi/cloudflare-warp/ipscanner"
)

const (
	defaultScanLimit   = 2
	defaultScanTimeout = time.Minute
	defaultMaxRTT      = time.Second
)

// ErrNoEndpoints is returned when a scan finds no usable WARP endpoint.
var ErrNoEndpoints = errors.New("no usable WARP endpoints found")

// Endpoint is a WARP server verified with a WireGuard handshake.
type Endpoint struct {
	AddrPort netip.AddrPort
	RTT      time.Duration
}

// ScanOptions controls WARP endpoint discovery.
type ScanOptions struct {
	// IPv4 and IPv6 select address families. When both are false, both are used.
	IPv4 bool
	IPv6 bool
	// Limit is the number of best endpoints to return. The default is 2.
	Limit int
	// MaxRTT rejects slower endpoints. The default is one second.
	MaxRTT time.Duration
	// Timeout limits the whole scan. The default is one minute. Partial results
	// are returned without an error when timeout expires after finding any.
	Timeout time.Duration
}

// Scan discovers WARP endpoints, verifies them with WireGuard handshakes, and
// returns results ordered from lowest to highest RTT.
func (c *Client) Scan(ctx context.Context, options ScanOptions) ([]Endpoint, error) {
	if c == nil || c.identity == nil {
		return nil, errors.New("WARP client is not initialized")
	}
	if ctx == nil {
		return nil, errors.New("context is nil")
	}

	options = normalizeScanOptions(options)
	scanCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	scanner := ipscanner.NewScanner(
		ipscanner.WithContext(scanCtx),
		ipscanner.WithWarpPrivateKey(c.identity.PrivateKey),
		ipscanner.WithWarpPeerPublicKey(c.identity.Config.Peers[0].PublicKey),
		ipscanner.WithUseIPv4(options.IPv4),
		ipscanner.WithUseIPv6(options.IPv6),
		ipscanner.WithMaxDesirableRTT(options.MaxRTT),
		ipscanner.WithIPQueueSize(options.Limit),
		ipscanner.WithCidrList(network.ScannerPrefixes()),
		ipscanner.WithCache(nil),
	)
	runDone := make(chan error, 1)
	go func() { runDone <- scanner.Run() }()
	defer scanner.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			scanner.Stop()
			<-scanner.Done()
			return nil, ctx.Err()
		case <-scanCtx.Done():
			return stopScanner(scanner, options.Limit)
		case err := <-runDone:
			if err != nil {
				return nil, err
			}
			return scanResults(scanner, options.Limit)
		case <-ticker.C:
			if len(scanner.GetAvailableIPs()) >= options.Limit {
				return stopScanner(scanner, options.Limit)
			}
		}
	}
}

func stopScanner(scanner *ipscanner.IPScanner, limit int) ([]Endpoint, error) {
	scanner.Stop()
	<-scanner.Done()
	return scanResults(scanner, limit)
}

func normalizeScanOptions(options ScanOptions) ScanOptions {
	if !options.IPv4 && !options.IPv6 {
		options.IPv4, options.IPv6 = true, true
	}
	if options.Limit <= 0 {
		options.Limit = defaultScanLimit
	}
	if options.MaxRTT <= 0 {
		options.MaxRTT = defaultMaxRTT
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultScanTimeout
	}
	return options
}

func scanResults(scanner *ipscanner.IPScanner, limit int) ([]Endpoint, error) {
	available := scanner.GetAvailableIPs()
	if len(available) == 0 {
		return nil, ErrNoEndpoints
	}
	if len(available) > limit {
		available = available[:limit]
	}
	result := make([]Endpoint, len(available))
	for index, info := range available {
		result[index] = Endpoint{AddrPort: info.AddrPort, RTT: info.RTT}
	}
	return result, nil
}

// CheckEndpoint verifies one exact IP and UDP port with a WARP WireGuard
// handshake and returns measured RTT.
func (c *Client) CheckEndpoint(ctx context.Context, endpoint netip.AddrPort) (Endpoint, error) {
	if c == nil || c.identity == nil {
		return Endpoint{}, errors.New("WARP client is not initialized")
	}
	info, err := ipscanner.CheckEndpoint(
		ctx,
		endpoint,
		c.identity.PrivateKey,
		c.identity.Config.Peers[0].PublicKey,
	)
	if err != nil {
		return Endpoint{}, err
	}
	return Endpoint{AddrPort: info.AddrPort, RTT: info.RTT}, nil
}
