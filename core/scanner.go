package core

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/network"
	"github.com/shahradelahi/cloudflare-warp/ipscanner"
	"github.com/shahradelahi/cloudflare-warp/log"
)

type ScanOptions struct {
	V4         bool
	V6         bool
	MaxRTT     time.Duration
	PrivateKey string
	PublicKey  string
}

func RunScan(ctx context.Context, opts ScanOptions) (result []ipscanner.IPInfo, err error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	scanner := ipscanner.NewScanner(
		ipscanner.WithWarpPrivateKey(opts.PrivateKey),
		ipscanner.WithWarpPeerPublicKey(opts.PublicKey),
		ipscanner.WithUseIPv4(opts.V4),
		ipscanner.WithUseIPv6(opts.V6),
		ipscanner.WithMaxDesirableRTT(opts.MaxRTT),
		ipscanner.WithCidrList(network.ScannerPrefixes()),
	)

	go func() {
		if err := scanner.Run(); err != nil {
			log.Errorw("IP scanner encountered a fatal error during execution", zap.Error(err))
		}
	}()
	defer scanner.Stop()

	startTime := time.Now()
	log.Info("Initiating IP scan process...")

	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

	checkTicker := time.NewTicker(250 * time.Millisecond)
	defer checkTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("scan canceled or timed out")
		case <-progressTicker.C:
			elapsed := time.Since(startTime).Round(time.Second)
			log.Infow("IP scan in progress", zap.Duration("elapsed_time", elapsed))
		case <-checkTicker.C:
			ipList := scanner.GetAvailableIPs()
			if len(ipList) > 1 {
				result = ipList[:2]
				elapsed := time.Since(startTime).Round(time.Second)
				log.Infow("IP scan completed successfully", zap.Int("endpoints_found", len(result)), zap.Duration("duration", elapsed))
				log.Debugw("Found endpoints", "endpoints", result)
				return result, nil
			}
		}
	}
}
