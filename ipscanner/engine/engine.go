package engine

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/ipscanner/ipgenerator"
	"github.com/shahradelahi/cloudflare-warp/ipscanner/model"
	"github.com/shahradelahi/cloudflare-warp/ipscanner/ping"
	"github.com/shahradelahi/cloudflare-warp/log"
)

type Engine struct {
	options    *statute.ScannerOptions
	generators []*ipgenerator.IpGenerator

	ipQueue *IPQueue
	ping    *ping.Ping

	ctx    context.Context
	cancel context.CancelFunc
}

func NewScannerEngine(ctx context.Context, opts *statute.ScannerOptions) (*Engine, error) {
	var generators []*ipgenerator.IpGenerator
	for _, cidr := range opts.CidrList {
		if !opts.UseIPv6 && cidr.Addr().Is6() {
			continue
		}
		if !opts.UseIPv4 && cidr.Addr().Is4() {
			continue
		}

		gen, err := ipgenerator.NewIpGenerator([]netip.Prefix{cidr})
		if err != nil {
			return nil, errors.New("failed to create IP generator")
		}
		generators = append(generators, gen)
	}

	childCtx, cancel := context.WithCancel(ctx)

	return &Engine{
		options:    opts,
		generators: generators,

		ipQueue: NewIPQueue(opts),

		ping:   ping.NewPinger(opts),
		ctx:    childCtx,
		cancel: cancel,
	}, nil
}

func (e *Engine) Run() {
	e.ipQueue.Init()

	processedIPs := 0
	progressTicker := time.NewTicker(5 * time.Second)

	defer progressTicker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			log.Info("Scanner Done")
			return
		case <-progressTicker.C:
			log.Infow("Scanning progress", zap.Int("processed_ips", processedIPs), zap.Int("found_ips", e.ipQueue.Size()))
		default:
			var wg sync.WaitGroup
			for _, generator := range e.generators {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ip, ok := generator.Next()
					if ok {
						e.pingAddr(ip)
					}
					processedIPs++
				}()
			}
			wg.Wait()
		}
	}
}

func (e *Engine) pingAddr(addr netip.Addr) {
	log.Debugw("Pinging IP", zap.String("ip", addr.String()))

	info, err := e.ping.DoPing(e.ctx, addr)
	if err != nil {
		log.Debugw("Ping failed", zap.String("ip", addr.String()), zap.Error(err))
		return
	}

	if e.options.Cache != nil {
		e.options.Cache.SaveEndpoint(info.AddrPort.String(), info.RTT)
	}

	if info.RTT < e.options.MaxDesirableRTT {
		e.ipQueue.Enqueue(info)
		log.Infow("Found desirable IP", zap.String("ip", info.AddrPort.String()), zap.Duration("rtt", info.RTT))
	} else {
		log.Debugw("IP pinged but RTT is too high", zap.String("ip", info.AddrPort.String()), zap.Duration("rtt", info.RTT))
	}
}

func (e *Engine) GetAvailableIPs(desc bool) []statute.IPInfo {
	if e.ipQueue != nil {
		return e.ipQueue.AvailableIPs(desc)
	}
	return nil
}

func (e *Engine) Shutdown() {
	e.cancel()
	e.ctx.Done()
}
