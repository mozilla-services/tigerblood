package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mozilla.org/tigerblood"
)

// reputationCmd represents the reputation command
var reputationCmd = &cobra.Command{
	Use:   "reputation",
	Short: "Request reputation for IP address.",
	Long:  `Requests the current reputation value for a given IP address.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires IP address argument")
		}
		if tigerblood.IsValidReputationIP(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid IP specified: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		ipaddr := args[0]
		url := viper.GetString("URL")

		client, err := tigerblood.NewClient(
			url,
			viper.GetString("HAWK_ID"),
			viper.GetString("HAWK_SECRET"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tigerblood client: %s\n", err)
			os.Exit(1)
		}

		resp, err := client.Reputation(ipaddr)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				fmt.Printf("reputation entry not found\n")
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Error requesting reputation: %s\n", err)
			os.Exit(1)
		}

		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response body: %s\n", err)
			os.Exit(1)
		}
		var r tigerblood.ReputationEntry
		err = json.Unmarshal(buf, &r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling response: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%v %v %v\n", r.IP, r.Reputation, r.Reviewed)
	},
}

func init() {
	rootCmd.AddCommand(reputationCmd)
}
