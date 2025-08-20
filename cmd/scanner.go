package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shahradelahi/cloudflare-warp/cloudflare"
	"github.com/shahradelahi/cloudflare-warp/cloudflare/network"
	"github.com/shahradelahi/cloudflare-warp/ipscanner"
	"github.com/shahradelahi/cloudflare-warp/log"
)

var ScannerCmd = &cobra.Command{
	Use:   "scanner",
	Short: "Scan for the best Cloudflare WARP IP",
	Long: `Scans for the best Cloudflare WARP IP addresses by testing a list of known CIDRs.
It measures the Round-Trip Time (RTT) and displays a list of the fastest available endpoints.
This is useful for finding optimal endpoints to use with the 'run' command for better performance.`,
	Run: runScanner,
}

func init() {
	ScannerCmd.Flags().BoolP("ipv4", "4", false, "Only scan for IPv4 WARP endpoints.")
	ScannerCmd.Flags().BoolP("ipv6", "6", false, "Only scan for IPv6 WARP endpoints.")
	ScannerCmd.Flags().Duration("rtt", 1000*time.Millisecond, "Maximum RTT (Round-Trip Time) for scanned IPs (e.g., 1000ms).")

	viper.BindPFlag("scanner.ipv4", ScannerCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("scanner.ipv6", ScannerCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("scanner.rtt", ScannerCmd.Flags().Lookup("rtt"))
}

func runScanner(cmd *cobra.Command, args []string) {
	v4, _ := cmd.Flags().GetBool("ipv4")
	v6, _ := cmd.Flags().GetBool("ipv6")
	rtt, _ := cmd.Flags().GetDuration("rtt")

	identity, err := cloudflare.LoadIdentity()
	if err != nil {
		fatal(fmt.Errorf("failed to load identity: %w", err))
	}

	// Essentially doing XNOR to make sure that if they are both false
	// or both true, just set them both true.
	if v4 == v6 {
		v4, v6 = true, true
	}

	// Create a context that is cancelled when an interrupt signal is received
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		log.Info("Interrupt signal received, stopping scanner...")
	}()

	log.Info("Starting IP scanning...")
	log.Info("Press CTRL+C to stop the scanner at any time.")

	// new scanner
	scanner := ipscanner.NewScanner(
		ipscanner.WithWarpPrivateKey(identity.PrivateKey),
		ipscanner.WithWarpPeerPublicKey(identity.Config.Peers[0].PublicKey),
		ipscanner.WithUseIPv4(v4),
		ipscanner.WithUseIPv6(v6),
		ipscanner.WithMaxDesirableRTT(rtt),
		ipscanner.WithCidrList(network.ScannerPrefixes()),
		ipscanner.WithIPQueueSize(0xffff),
		ipscanner.WithContext(ctx),
	)

	if err := scanner.Run(); err != nil {
		if !errors.Is(err, context.Canceled) {
			fatal(err)
		}
	}

	log.Info("IP scanning process completed.")

	ipList := scanner.GetAvailableIPs()

	if len(ipList) == 0 {
		log.Info("No desirable IP endpoints were found during the scan.")
		return
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Address", "RTT (ping)", "Time")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, info := range ipList {
		tbl.AddRow(info.AddrPort, info.RTT, info.CreatedAt.Format(time.DateTime))
	}

	tbl.Print()
}
