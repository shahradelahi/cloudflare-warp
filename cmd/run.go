package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/core"
	"github.com/shahradelahi/cloudflare-warp/core/cache"
	"github.com/shahradelahi/cloudflare-warp/ipscanner/ipgenerator"
	"github.com/shahradelahi/cloudflare-warp/log"
	"github.com/shahradelahi/cloudflare-warp/utils"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Cloudflare WARP proxy",
	Long: `Run the Cloudflare WARP proxy.
This command starts the proxy server and establishes a connection to the Cloudflare network.
You can configure the proxy to use Socks5 or HTTP, specify WARP endpoints, and enable features like WARP+ connections.`,
	Run: run,
}

func init() {
	RunCmd.Flags().Bool("4", false, "Use IPv4 for random WARP endpoint selection.")
	RunCmd.Flags().Bool("6", false, "Use IPv6 for random WARP endpoint selection.")
	RunCmd.Flags().String("socks-addr", "", "Socks5 proxy bind address.")
	RunCmd.Flags().String("http-addr", "", "HTTP proxy bind address.")
	RunCmd.Flags().String("dns", "1.1.1.1", "DNS server address to use (e.g., 1.1.1.1).")
	RunCmd.Flags().StringSliceP("endpoint", "e", []string{}, "Specify a custom WARP endpoint.")
	RunCmd.Flags().Bool("scan", false, "Enable WARP IP scanning before connecting.")
	RunCmd.Flags().Duration("scan-rtt", 1000*time.Millisecond, "Scanner RTT limit for endpoint selection (e.g., 1000ms).")

	viper.BindPFlag("4", RunCmd.Flags().Lookup("4"))
	viper.BindPFlag("6", RunCmd.Flags().Lookup("6"))
	viper.BindPFlag("socks-addr", RunCmd.Flags().Lookup("socks-addr"))
	viper.BindPFlag("http-addr", RunCmd.Flags().Lookup("http-addr"))
	viper.BindPFlag("endpoint", RunCmd.Flags().Lookup("endpoint"))
	viper.BindPFlag("dns", RunCmd.Flags().Lookup("dns"))
	viper.BindPFlag("scan", RunCmd.Flags().Lookup("scan"))
	viper.BindPFlag("scan-rtt", RunCmd.Flags().Lookup("scan-rtt"))
}

func run(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if viper.GetBool("4") && viper.GetBool("6") {
		fatal(errors.New("can't force v4 and v6 at the same time"))
	}

	socksAddrStr := viper.GetString("socks-addr")
	httpAddrStr := viper.GetString("http-addr")

	if socksAddrStr == "" && httpAddrStr == "" {
		// if no flags are set, print help
		cmd.Help()
		return
	}

	useV4, useV6 := viper.GetBool("4"), viper.GetBool("6")
	if !useV4 && !useV6 {
		useV4, useV6 = true, true
	}

	var socksAddr, httpAddr *netip.AddrPort
	if socksAddrStr != "" {
		addr, err := netip.ParseAddrPort(socksAddrStr)
		if err != nil {
			fatal(fmt.Errorf("invalid Socks5 bind address: %w", err))
		}
		socksAddr = &addr
	}

	if httpAddrStr != "" {
		addr, err := netip.ParseAddrPort(httpAddrStr)
		if err != nil {
			fatal(fmt.Errorf("invalid HTTP bind address: %w", err))
		}
		httpAddr = &addr
	}

	dnsAddr, err := netip.ParseAddr(viper.GetString("dns"))
	if err != nil {
		fatal(fmt.Errorf("invalid DNS address: %w", err))
	}

	endpoints := viper.GetStringSlice("endpoint")
	userProvidedEndpoint := len(endpoints) > 0

	opts := core.Config{
		SocksBindAddress:     socksAddr,
		HttpBindAddress:      httpAddr,
		Endpoints:            endpoints,
		DnsAddr:              dnsAddr,
		UserProvidedEndpoint: userProvidedEndpoint,
	}

	c := cache.NewCache()

	if viper.GetBool("scan") {
		log.Infow("Scanner mode enabled", zap.Duration("max-rtt", viper.GetDuration("rtt")))
		opts.Scan = &core.ScanOptions{
			V4:     useV4,
			V6:     useV6,
			MaxRTT: viper.GetDuration("rtt"),
		}
	}

	if len(opts.Endpoints) == 0 && !viper.GetBool("scan") {
		endpoints, err := c.GetRandomEndpoints(1)
		if err != nil {
			addr, err := utils.ParseResolveAddressPort("engage.cloudflareclient.com:2408", false, opts.DnsAddr.String())
			if err == nil {
				iprange, _ := ipgenerator.NewIPRange(netip.PrefixFrom(addr.Addr(), 24))
				ips := iprange.GetAll()
				opts.Endpoints = []string{
					fmt.Sprintf("%s:2408", ips[0]),
					fmt.Sprintf("%s:500", ips[1]),
				}
			} else {
				log.Warnw("Not enough available endpoints found in cache; automatically enabling scanner mode to discover new endpoints.")
				opts.Scan = &core.ScanOptions{
					V4:     useV4,
					V6:     useV6,
					MaxRTT: viper.GetDuration("rtt"),
				}
			}
		} else {
			opts.Endpoints = endpoints
			log.Infow("Using random endpoint from cache", zap.String("endpoint", opts.Endpoints[0]))
		}
	}

	engine := core.NewEngine(ctx, opts)
	defer engine.Stop()

	go func() {
		if err := engine.Run(); err != nil {
			fatal(err)
		}
	}()

	<-ctx.Done()
}

func fatal(err error) {
	log.Fatalw("Application encountered a fatal error", zap.Error(err))
}
