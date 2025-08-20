package ipscanner

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/core/cache"
	"github.com/shahradelahi/cloudflare-warp/ipscanner/engine"
	"github.com/shahradelahi/cloudflare-warp/ipscanner/model"
	"github.com/shahradelahi/cloudflare-warp/log"
)

type IPScanner struct {
	options statute.ScannerOptions
	engine  *engine.Engine
	ctx     context.Context
}

func NewScanner(options ...Option) *IPScanner {
	// Create a new cache instance
	c := cache.NewCache()

	// Load the cache from a file
	if err := c.LoadCache(); err != nil {
		log.Warnw("Failed to load existing IP scan cache; starting with an empty cache", zap.Error(err))
	}

	p := &IPScanner{
		options: statute.ScannerOptions{
			UseIPv4:           true,
			UseIPv6:           true,
			CidrList:          statute.DefaultCFRanges(),
			WarpPresharedKey:  "",
			WarpPeerPublicKey: "",
			WarpPrivateKey:    "",
			IPQueueSize:       8,
			MaxDesirableRTT:   400 * time.Millisecond,
			IPQueueTTL:        30 * time.Second,
			Cache:             c,
		},
		ctx: context.Background(),
	}

	for _, option := range options {
		option(p)
	}

	return p
}

type Option func(*IPScanner)

func WithContext(ctx context.Context) Option {
	return func(i *IPScanner) {
		i.ctx = ctx
	}
}

func WithUseIPv4(useIPv4 bool) Option {
	return func(i *IPScanner) {
		i.options.UseIPv4 = useIPv4
	}
}

func WithUseIPv6(useIPv6 bool) Option {
	return func(i *IPScanner) {
		i.options.UseIPv6 = useIPv6
	}
}

func WithCidrList(cidrList []netip.Prefix) Option {
	return func(i *IPScanner) {
		i.options.CidrList = cidrList
	}
}

func WithIPQueueSize(size int) Option {
	return func(i *IPScanner) {
		i.options.IPQueueSize = size
	}
}

func WithMaxDesirableRTT(threshold time.Duration) Option {
	return func(i *IPScanner) {
		i.options.MaxDesirableRTT = threshold
	}
}

func WithIPQueueTTL(ttl time.Duration) Option {
	return func(i *IPScanner) {
		i.options.IPQueueTTL = ttl
	}
}

func WithWarpPrivateKey(privateKey string) Option {
	return func(i *IPScanner) {
		i.options.WarpPrivateKey = privateKey
	}
}

func WithWarpPeerPublicKey(peerPublicKey string) Option {
	return func(i *IPScanner) {
		i.options.WarpPeerPublicKey = peerPublicKey
	}
}

func WithWarpPreSharedKey(presharedKey string) Option {
	return func(i *IPScanner) {
		i.options.WarpPresharedKey = presharedKey
	}
}

func WithCache(c *cache.Cache) Option {
	return func(i *IPScanner) {
		i.options.Cache = c
	}
}

// run engine and in case of new event call onChange callback also if it gets canceled with context
// cancel all operations

func (i *IPScanner) Run() error {
	if !i.options.UseIPv4 && !i.options.UseIPv6 {
		log.Fatal("Invalid configuration: Both IPv4 and IPv6 scanning are disabled. Please enable at least one to proceed.")
		return nil
	}

	eng, err := engine.NewScannerEngine(i.ctx, &i.options)
	if err != nil {
		return errors.New("failed to create scanner engine")
	}

	i.engine = eng
	i.engine.Run()

	if i.options.Cache != nil {
		// Save the cache to a file
		if err := i.options.Cache.SaveCache(); err != nil {
			log.Warnw("Failed to save IP scan results to cache file", zap.Error(err))
		}
	}

	return nil
}

func (i *IPScanner) Stop() {
	if i.engine != nil {
		i.engine.Shutdown()
	}
}

func (i *IPScanner) GetAvailableIPs() []statute.IPInfo {
	if i.engine != nil {
		return i.engine.GetAvailableIPs(false)
	}
	return nil
}

type IPInfo = statute.IPInfo
