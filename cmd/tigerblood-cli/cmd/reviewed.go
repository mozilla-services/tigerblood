package cmd

import (
	"errors"
	"fmt"
	"net/http"
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
			fmt.Printf("Error creating tigerblood client:\n%s\n", err)
			os.Exit(1)
		}

		resp, err := client.SetReviewed(ipaddr, flag)
		if err != nil {
			fmt.Printf("Error setting reviewed flag:\n%s\n", err)
			os.Exit(1)
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Bad response setting reviewed flag:\n%s\n", resp)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(reviewedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// banCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// banCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
