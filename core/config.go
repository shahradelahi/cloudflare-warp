package core

import (
	"net/netip"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
)

// Config holds the configuration for the WARP engine.
type Config struct {
	SocksBindAddress     *netip.AddrPort
	HttpBindAddress      *netip.AddrPort
	Endpoints            []string
	DnsAddr              netip.Addr
	Scan                 *ScanOptions
	UserProvidedEndpoint bool
	// Identity overrides identity loading from the global data directory.
	// It allows embedded applications to reuse an identity across proxies.
	Identity *model.Identity
}
