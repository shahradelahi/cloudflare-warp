package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/cloudflare"
	"github.com/shahradelahi/cloudflare-warp/log"
)

var updateShortMsg = "Updates the Cloudflare Warp device configuration"

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: updateShortMsg,
	Long:  FormatMessage(updateShortMsg, `Updates device name and/or license key.`),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUpdate(); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	UpdateCmd.Flags().StringP("name", "n", "", "The new device name")
	UpdateCmd.Flags().StringP("license", "k", "", "The new license key")
	viper.BindPFlag("update.name", UpdateCmd.Flags().Lookup("name"))
	viper.BindPFlag("update.license", UpdateCmd.Flags().Lookup("license"))
}

func runUpdate() error {
	name := viper.GetString("update.name")
	license := viper.GetString("update.license")

	if name == "" && license == "" {
		return fmt.Errorf("at least one of --name or --license must be provided")
	}

	identity, err := cloudflare.LoadIdentity()
	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, errors.New("identity contains 0 peers")) {
			return fmt.Errorf("WARP identity not found. Please run 'warp generate' to create one")
		}
		return err
	}

	warpAPI := cloudflare.NewWarpAPI()
	updated := false

	// Update device name if provided
	if name != "" {
		_, err = warpAPI.UpdateSourceDevice(identity.Token, identity.ID, map[string]interface{}{"name": name})
		if err != nil {
			return fmt.Errorf("failed to update device name: %w", err)
		}
		fmt.Println("Device name updated successfully.")
		updated = true
	}

	// Update license if provided
	if license != "" {
		// Generate configs
		identity, err = cloudflare.CreateOrUpdateIdentity(license)
		if err != nil {
			log.Fatalw("Failed to generate primary identity", zap.Error(err))
		}

		fmt.Println("License updated successfully.")
		updated = true
	}

	// Save the updated identity to conf.json
	if updated {
		if err := identity.SaveIdentity(); err != nil {
			return fmt.Errorf("failed to save updated configuration: %w", err)
		}
		fmt.Println("Local configuration files updated.")
	}

	return nil
}
