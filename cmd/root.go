package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shahradelahi/cloudflare-warp/core/cache"
	"github.com/shahradelahi/cloudflare-warp/core/datadir"
	"github.com/shahradelahi/cloudflare-warp/internal/version"
	"github.com/shahradelahi/cloudflare-warp/log"
)

var rootCmd = &cobra.Command{
	Use:   "warp",
	Short: "Warp is an open-source implementation of Cloudflare's Warp.",
	Long: `Warp is an open-source implementation of Cloudflare's Warp client that allows you to route your internet traffic through Cloudflare's network, enhancing privacy and security.
It can operate as a Socks5 or HTTP proxy.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logLevel, err := log.ParseLevel(viper.GetString("loglevel"))
		if err != nil {
			panic(err)
		}
		log.SetLogger(log.Must(log.NewLeveled(logLevel)))

		// Initialize data directory
		dir := datadir.GetDataDirOrPath(viper.GetString("data-dir"))
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create data directory: %v\n", err)
			os.Exit(1)
		}
		datadir.SetDataDir(dir)

		// Load cache
		c := cache.NewCache()
		if err := c.LoadCache(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to load existing endpoint cache; starting with an empty cache: %v\n", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			fmt.Println(version.String())
			fmt.Println(version.BuildString())
			os.Exit(0)
		}
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("data-dir", "", "Directory to store generated profiles and identity files.")
	rootCmd.PersistentFlags().String("loglevel", "info", "SetDataDir the logging level (debug, info, warn, error, silent).")
	rootCmd.Flags().Bool("version", false, "Display version number.")

	viper.BindPFlag("data-dir", rootCmd.PersistentFlags().Lookup("data-dir"))
	viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("version", rootCmd.Flags().Lookup("version"))

	// Add subcommands
	rootCmd.AddCommand(RunCmd)
	rootCmd.AddCommand(ScannerCmd)
	rootCmd.AddCommand(GenerateCmd)
	rootCmd.AddCommand(StatusCmd)
	rootCmd.AddCommand(UpdateCmd)
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "Error reading config file:", err)
		}
	}
}
