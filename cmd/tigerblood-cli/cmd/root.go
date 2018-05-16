package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tigerblood-cli",
	Short: "Command line client for managing IP Reputations",
	Long: `Command line client for managing IP Reputations. It requires the environment variables
TIGERBLOOD_HAWK_ID, TIGERBLOOD_HAWK_SECRET, TIGERBLOOD_URL to be set, or a valid config file.

Example usage:

TIGERBLOOD_HAWK_ID=root TIGERBLOOD_HAWK_SECRET=toor TIGERBLOOD_URL=http://localhost:8080/ tigerblood-cli ban 192.8.8.0
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	viper.SetDefault("HAWK_ID", nil)
	viper.SetDefault("HAWK_SECRET", nil)
	viper.SetDefault("URL", "https://tigerblood.stage.mozaws.net/")

	viper.SetEnvPrefix("tigerblood")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.tigerblood-cli.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %s\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tigerblood-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tigerblood-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
			os.Exit(1)
		}
	}
}
