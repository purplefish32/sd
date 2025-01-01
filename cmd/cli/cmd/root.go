package cmd

import (
	"os"
	"sd/cmd/cli/cmd/button"
	"sd/cmd/cli/cmd/device"
	"sd/cmd/cli/cmd/page"
	"sd/cmd/cli/cmd/profile"
	i "sd/pkg/instance"

	"sd/cmd/cli/cmd/instance"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "sd",
	Short: "sd is a powerful way to control your Stream Deck via the command line",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("current-instance", i.GetInstanceId())
		viper.SetDefault("current-device", "")

		// Load the configuration
		viper.SetConfigName("config")    // name of config file (without extension)
		viper.SetConfigType("json")      // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath("$HOME/.sd") // Add the search path

		// Attempt to read the configuration file
		if err := viper.ReadInConfig(); err != nil {
			log.Error().Err(err).Msg("Failed to read config")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(instance.NewCmd())
	rootCmd.AddCommand(device.NewCmd())
	rootCmd.AddCommand(page.NewCmd())
	rootCmd.AddCommand(profile.NewCmd())
	rootCmd.AddCommand(button.NewCmd())
}
