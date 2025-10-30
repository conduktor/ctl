package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/client"
	"github.com/conduktor/ctl/internal/utils"
	"github.com/spf13/cobra"
)

// loginCmd represents the apply command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login user using username password to get a JWT token",
	Long:  `Use must use CDK_USER CDK_PASSWORD environment variables to login`,
	Args:  cobra.RangeArgs(0, 0),
	Run: func(cmd *cobra.Command, args []string) {
		specificAPIClient, err := client.Make(client.APIParameter{BaseURL: os.Getenv("CDK_BASE_URL"), Debug: utils.CdkDebug()})
		if *debug {
			specificAPIClient.ActivateDebug()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not login: %s\n", err)
			os.Exit(1)
		}
		username := os.Getenv("CDK_USER")
		if username == "" {
			fmt.Fprintln(os.Stderr, "Please set CDK_USER")
			os.Exit(2)
		}
		password := os.Getenv("CDK_PASSWORD")
		if password == "" {
			fmt.Fprintln(os.Stderr, "Please set CDK_PASSWORD")
			os.Exit(3)
		}
		token, err := specificAPIClient.Login(username, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not login: %s\n", err)
			os.Exit(4)
		}
		fmt.Println(token.AccessToken)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
