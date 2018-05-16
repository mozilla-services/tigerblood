package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mozilla.org/tigerblood"
)

// unbanCmd represents the unban command
var unbanCmd = &cobra.Command{
	Use:   "unban",
	Short: "Sets the reputation for an IPv4 CIDR to the maximum (100) to unban an IP.",
	Long:  `Sets the reputation for an IPv4 CIDR to 100.`,
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

		_, err = client.UnbanIP(cidr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unbanning IP: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Unbanned %s on %s\n", cidr, url)
	},
}

func init() {
	rootCmd.AddCommand(unbanCmd)
}
