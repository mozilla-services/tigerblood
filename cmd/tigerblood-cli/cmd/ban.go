package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mozilla.org/tigerblood"
)

// banCmd represents the ban command
var banCmd = &cobra.Command{
	Use:   "ban",
	Short: "Ban an IP for the maximum decay period (environment dependent).",
	Long:  `Sets the reputation for an IPv4 CIDR to 0.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one CIDR")
		}
		if tigerblood.IsValidReputationCIDROrIP(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid CIDR specified: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		cidr := args[0]
		url := viper.GetString("URL")

		client, err := tigerblood.NewClient(
			url,
			viper.GetString("HAWK_ID"),
			viper.GetString("HAWK_SECRET"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tigerblood client: %s\n", err)
			os.Exit(1)
		}

		_, err = client.BanIP(cidr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error banning IP: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Banned %s on %s\n", cidr, url)
	},
}

func init() {
	rootCmd.AddCommand(banCmd)
}
