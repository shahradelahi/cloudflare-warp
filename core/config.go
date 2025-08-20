package core

import (
	"net/netip"
)

// Config holds the configuration for the WARP engine.
type Config struct {
	SocksBindAddress     *netip.AddrPort
	HttpBindAddress      *netip.AddrPort
	Endpoints            []string
	DnsAddr              netip.Addr
	Scan                 *ScanOptions
	UserProvidedEndpoint bool
}
