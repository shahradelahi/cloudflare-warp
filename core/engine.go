package core

import (
	"context"
	"errors"
	"net/netip"

	"github.com/shahradelahi/wiresocks"
	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/cloudflare"
	cache2 "github.com/shahradelahi/cloudflare-warp/core/cache"
	"github.com/shahradelahi/cloudflare-warp/log"
)

// Engine is the main engine for running WARP.
type Engine struct {
	ctx    context.Context
	opts   Config
	cancel context.CancelFunc
	cache  *cache2.Cache
}

// NewEngine creates a new WARP engine.
func NewEngine(ctx context.Context, opts Config) *Engine {
	ctx, cancel := context.WithCancel(ctx)
	return &Engine{
		ctx:    ctx,
		opts:   opts,
		cancel: cancel,
		cache:  cache2.NewCache(),
	}
}

// Run runs the WARP engine.
func (e *Engine) Run() error {
	var endpoints []string

	if e.opts.Scan != nil {
		scannedEndpoints, err := e.getScannerEndpoints()
		if err != nil {
			return err
		}
		endpoints = scannedEndpoints
	} else {
		endpoints = e.opts.Endpoints
	}

	for {
		select {
		case <-e.ctx.Done():
			return e.ctx.Err()
		default:
			if len(endpoints) < 1 && !e.opts.UserProvidedEndpoint {
				newEndpoint, _ := e.cache.GetRandomEndpoints(1)
				if newEndpoint != nil {
					endpoints = newEndpoint
					log.Infow("Using new random endpoints from cache", zap.Strings("endpoints", endpoints))
				}
			}

			if len(endpoints) == 0 {
				return errors.New("no endpoint available")
			}

			log.Infow("Connecting to WARP endpoint(s)", zap.Strings("endpoints", endpoints))

			if err := e.runWarp(endpoints[0]); err != nil {
				log.Errorw("WARP connection failed", zap.Error(err), zap.Strings("endpoints", endpoints))
				endpoints = endpoints[1:]
				if e.opts.UserProvidedEndpoint {
					return err
				}
			} else {
				// connection successful, wait for context to be done
				<-e.ctx.Done()
				return nil
			}
		}
	}
}

func (e *Engine) getScannerEndpoints() ([]string, error) {
	// make primary identity
	ident, err := cloudflare.LoadOrCreateIdentity()
	if err != nil {
		log.Errorw("Failed to load/create primary identity", zap.Error(err))
		return nil, err
	}

	// Reading the private key from the 'Interface' section
	e.opts.Scan.PrivateKey = ident.PrivateKey

	// Reading the public key from the 'Peer' section
	e.opts.Scan.PublicKey = ident.Config.Peers[0].PublicKey

	res, err := RunScan(e.ctx, *e.opts.Scan)
	if err != nil {
		return nil, err
	}

	endpointStrings := make([]string, len(res))
	for i, ipInfo := range res {
		endpointStrings[i] = ipInfo.AddrPort.String() + " (RTT: " + ipInfo.RTT.String() + ")"
	}
	log.Infow("Scan successful", "found", len(res))

	endpoints := make([]string, len(res))
	for i := 0; i < len(res); i++ {
		endpoints[i] = res[i].AddrPort.String()
	}
	return endpoints, nil
}

// Stop stops the WARP engine.
func (e *Engine) Stop() {
	e.cancel()
}

func (e *Engine) runWarp(endpoint string) error {
	// make primary identity
	ident, err := cloudflare.LoadOrCreateIdentity()
	if err != nil {
		log.Errorw("Failed to load primary identity", zap.Error(err))
		return err
	}

	conf := GenerateWireguardConfig(ident)

	// Set up DNS Address
	conf.Interface.DNS = []netip.Addr{e.opts.DnsAddr}

	// Enable keepalive on all peers in config
	for i, peer := range conf.Peers {
		peer.Endpoint = endpoint
		peer.KeepAlive = 5

		conf.Peers[i] = peer
	}

	proxyOpts := wiresocks.ProxyConfig{
		SocksBindAddr: e.opts.SocksBindAddress,
		HttpBindAddr:  e.opts.HttpBindAddress,
	}

	return e.startProxy(e.ctx, &conf, &proxyOpts)
}

// startProxy starts the proxy servers and waits for the context to be done.
func (e *Engine) startProxy(ctx context.Context, conf *wiresocks.Configuration, opts *wiresocks.ProxyConfig) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	ws, err := wiresocks.NewWireSocks(
		wiresocks.WithContext(ctx),
		wiresocks.WithWireguardConfig(conf),
		wiresocks.WithProxyConfig(opts),
	)
	if err != nil {
		return err
	}

	go func() {
		if err := ws.Run(); err != nil {
			log.Errorw("Failed to to start proxy server", zap.Error(err))
			cancel(err)
		}
	}()

	if opts.SocksBindAddr != nil {
		log.Infow("Serving Socks5 proxy", zap.Stringer("addr", opts.SocksBindAddr))
	}
	if opts.HttpBindAddr != nil {
		log.Infow("Serving HTTP proxy", zap.Stringer("addr", opts.HttpBindAddr))
	}

	<-ctx.Done()
	return nil
}
