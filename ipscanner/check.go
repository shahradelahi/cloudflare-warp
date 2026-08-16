package ipscanner

import (
	"context"
	"errors"
	"net/netip"

	"github.com/shahradelahi/cloudflare-warp/ipscanner/model"
	"github.com/shahradelahi/cloudflare-warp/ipscanner/ping"
)

// CheckEndpoint performs a WireGuard handshake with one exact WARP endpoint.
func CheckEndpoint(ctx context.Context, endpoint netip.AddrPort, privateKey, peerPublicKey string) (IPInfo, error) {
	if ctx == nil {
		return IPInfo{}, errors.New("context is nil")
	}
	if !endpoint.IsValid() || endpoint.Port() == 0 {
		return IPInfo{}, errors.New("invalid WARP endpoint")
	}
	if privateKey == "" || peerPublicKey == "" {
		return IPInfo{}, errors.New("WARP keys are required")
	}

	options := &statute.ScannerOptions{
		WarpPrivateKey:    privateKey,
		WarpPeerPublicKey: peerPublicKey,
		ScannerPorts:      []uint16{endpoint.Port()},
	}
	return ping.NewPinger(options).DoPing(ctx, endpoint.Addr())
}
