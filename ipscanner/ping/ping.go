package ping

import (
	"context"
	"net/netip"

	"github.com/shahradelahi/cloudflare-warp/ipscanner/model"
)

type Ping struct {
	options *statute.ScannerOptions
}

// NewPinger creates a new Ping instance.
func NewPinger(opts *statute.ScannerOptions) *Ping {
	return &Ping{
		options: opts,
	}
}

// DoPing performs a ping on the given IP address.
func (p *Ping) DoPing(ctx context.Context, ip netip.Addr) (statute.IPInfo, error) {
	res, err := p.calc(ctx, NewWarpPing(ip, p.options))
	if err != nil {
		return statute.IPInfo{}, err
	}

	return res, nil
}

func (p *Ping) calc(ctx context.Context, tp statute.IPing) (statute.IPInfo, error) {
	pr := tp.PingContext(ctx)
	err := pr.Error()
	if err != nil {
		return statute.IPInfo{}, err
	}
	return pr.Result(), nil
}
