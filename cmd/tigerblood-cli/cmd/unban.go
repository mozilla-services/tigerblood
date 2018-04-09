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
			fmt.Printf("Error creating tigerblood client:\n%s\n", err)
			os.Exit(1)
		}

		resp, err := client.UnbanIP(cidr)
		if err != nil {
			fmt.Printf("Error unbanning IP:\n%s\n", err)
			os.Exit(1)
		} else if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			fmt.Printf("Unbanned %s on %s\n", cidr, url)
		} else {
			fmt.Printf("Bad response unbanning IP:\n%s\n", resp)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(unbanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unbanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unbanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
