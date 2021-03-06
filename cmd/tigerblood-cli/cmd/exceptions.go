package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mozilla.org/tigerblood"
)

// exceptionsCmd represents the exceptions command
var exceptionsCmd = &cobra.Command{
	Use:   "exceptions",
	Short: "Display current exceptions list.",
	Long:  `Request and display current tigerblood exception list.`,
	Run: func(cmd *cobra.Command, args []string) {
		url := viper.GetString("URL")

		client, err := tigerblood.NewClient(
			url,
			viper.GetString("HAWK_ID"),
			viper.GetString("HAWK_SECRET"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tigerblood client: %s\n", err)
			os.Exit(1)
		}

		resp, err := client.Exceptions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error requesting exceptions: %s\n", err)
			os.Exit(1)
		}

		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response body: %s\n", err)
			os.Exit(1)
		}
		var e []tigerblood.ExceptionEntry
		err = json.Unmarshal(buf, &e)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling response: %s\n", err)
			os.Exit(1)
		}
		for _, x := range e {
			fmt.Printf("%v %v %v %v\n", x.IP, x.Creator, x.Modified, x.Expires)
		}
	},
}

func init() {
	rootCmd.AddCommand(exceptionsCmd)
}
