package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/shahradelahi/cloudflare-warp/cloudflare"
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
		if os.IsNotExist(err) || errors.Is(err, errors.New("identity contains 0 peers")) {
			return fmt.Errorf("WARP identity not found. Please run 'warp generate' to create one")
		}
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
