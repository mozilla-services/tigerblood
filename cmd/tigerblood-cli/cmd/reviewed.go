package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mozilla.org/tigerblood"
)

// reviewedCmd represents the reviewed command
var reviewedCmd = &cobra.Command{
	Use:   "reviewed",
	Short: "Change reviewed status.",
	Long:  `Set the reviewed status for a given reputation entry.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires CIDR and true or false")
		}
		if !tigerblood.IsValidReputationCIDROrIP(args[0]) {
			return fmt.Errorf("invalid CIDR specified: %s", args[0])
		}
		if strings.ToLower(args[1]) != "true" && strings.ToLower(args[1]) != "false" {
			return fmt.Errorf("reviewed status must be true or false")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ipaddr := args[0]
		flag := false
		if strings.ToLower(args[1]) == "true" {
			flag = true
		}
		url := viper.GetString("URL")

		client, err := tigerblood.NewClient(
			url,
			viper.GetString("HAWK_ID"),
			viper.GetString("HAWK_SECRET"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tigerblood client: %s\n", err)
			os.Exit(1)
		}

		_, err = client.SetReviewed(ipaddr, flag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting reviewed flag: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Flag for %s set to %t\n", ipaddr, flag)
	},
}

func init() {
	rootCmd.AddCommand(reviewedCmd)
}
