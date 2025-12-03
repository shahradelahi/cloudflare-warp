package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/cloudflare"
	"github.com/shahradelahi/cloudflare-warp/core"
	"github.com/shahradelahi/cloudflare-warp/log"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates and prints the WireGuard configuration",
	Long:  `This command generates and prints the WireGuard configuration based on your WARP identity. The output can be redirected to a file to create a WireGuard configuration file.`,
	Run:   generate,
}

func init() {}

func generate(cmd *cobra.Command, args []string) {
	ident, err := cloudflare.LoadOrCreateIdentity()
	if err != nil {
		log.Fatalw("Failed to generate primary identity", zap.Error(err))
	}

	wgConf := core.GenerateWireguardConfig(ident)

	confStr, err := wgConf.String()
	if err != nil {
		log.Fatalw("Failed to generate WireGuard configuration", zap.Error(err))
	}

	fmt.Println(confStr)
}
