package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/utils"
	"os"

	"github.com/spf13/cobra"
)

// consoleUrlGetCmd represents the apply command
var consoleUrlGetCmd = &cobra.Command{
	Use:   "consoleUrlGet",
	Short: "Perform a get on the given url/path with correct authentication header",
	Long: `Perform a get on the given url/path with correct authentication header.
If a full url is given the host part will be ignored and replace by $CDK_BASE_URL. If /api is missing it will be added automatically`,
	Args:   cobra.ExactArgs(1),
	Hidden: !utils.CdkDevMode(),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := consoleApiClient().UrlGet(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not perform get: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(consoleUrlGetCmd)
}
