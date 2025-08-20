package cmd

import (
	"log"

	"github.com/shahradelahi/cloudflare-warp/cloudflare"
	"github.com/spf13/cobra"
)

var statusShortMsg = "Prints the status of the current Cloudflare Warp device"

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: statusShortMsg,
	Long:  FormatMessage(statusShortMsg, ``),
	Run: func(cmd *cobra.Command, args []string) {
		if err := status(); err != nil {
			log.Fatal(err)
		}
	},
}

func status() error {
	identity, err := cloudflare.LoadIdentity()
	if err != nil {
		return err
	}

	warpAPI := cloudflare.NewWarpAPI()

	thisDevice, err := warpAPI.GetSourceDevice(identity.Token, identity.ID)
	if err != nil {
		return err
	}

	boundDevice, err := warpAPI.GetSourceBoundDevice(identity.Token, identity.ID)
	if err != nil {
		return err
	}

	PrintDeviceData(&thisDevice, boundDevice)
	return nil
}
