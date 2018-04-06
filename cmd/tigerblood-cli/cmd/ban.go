package cmd

import (
	"errors"
	"fmt"
	"net/http"
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
			fmt.Printf("Error creating tigerblood client:\n%s\n", err)
			os.Exit(1)
		}

		resp, err := client.BanIP(cidr)
		if err != nil {
			fmt.Printf("Error banning IP:\n%s\n", err)
			os.Exit(1)
		} else if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			fmt.Printf("Banned %s on %s\n", cidr, url)
		} else {
			fmt.Printf("Bad response banning IP:\n%s\n", resp)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(banCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// banCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// banCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
