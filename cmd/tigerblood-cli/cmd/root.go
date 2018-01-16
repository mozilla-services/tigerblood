
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
	Long: `Command line client for managing IP Reputations. It requires
the environment variables TIGERBLOOD_HAWK_ID, TIGERBLOOD_HAWK_SECRET, TIGERBLOOD_URL. Example usage:

TIGERBLOOD_HAWK_ID=root TIGERBLOOD_HAWK_SECRET=toor TIGERBLOOD_URL=http://localhost:8000/ tigerblood-cli ban 192.8.8.0/8
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	viper.SetDefault("HAWK_ID", nil)
	viper.SetDefault("HAWK_SECRET", nil)
	viper.SetDefault("URL", "https://tigerblood.stage.mozaws.net/")

	viper.SetEnvPrefix("tigerblood")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tigerblood-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tigerblood-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tigerblood-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
