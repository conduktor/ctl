package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var applicationInstanceNameForList *string
var applicationInstanceNameForCreate *string
var applicationNameForList *string
var applicationNameForCreate *string
var expiresAtForCreateAdmin *string
var expiresAtForCreateApplication *string

func validateExpiresAt(expiresAt string) {
	if expiresAt == "" {
		return
	}
	if _, err := time.Parse(time.RFC3339, expiresAt); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid --expires-at %q: expected RFC3339, e.g. 2027-01-01T00:00:00Z\n", expiresAt)
		os.Exit(1)
	}
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage Admin, Application and Application Instance tokens",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(1)
	},
}

var listTokenCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(1)
	},
}

var listAdminCmd = &cobra.Command{
	Use:  "admin",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := consoleAPIClient().ListAdminToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not list admin token: %s\n", err)
			os.Exit(1)
		}
		if len(result) == 0 {
			fmt.Println("No tokens found")
		}
		for _, token := range result {
			fmt.Printf("%s:\t%s\n", token.Name, token.Id)
		}
	},
}

var listApplicationInstanceTokenCmd = &cobra.Command{
	Use:  "application-instance",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := consoleAPIClient().ListApplicationInstanceToken(*applicationInstanceNameForList)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not list application-instance token: %s\n", err)
			os.Exit(1)
		}
		if len(result) == 0 {
			fmt.Println("No tokens found")
		}
		for _, token := range result {
			fmt.Printf("%s:\t%s\n", token.Name, token.Id)
		}
	},
}

var listApplicationTokenCmd = &cobra.Command{
	Use:  "application",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := consoleAPIClient().ListApplicationToken(*applicationNameForList)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not list application token: %s\n", err)
			os.Exit(1)
		}
		if len(result) == 0 {
			fmt.Println("No tokens found")
		}
		for _, token := range result {
			line := fmt.Sprintf("%s:\t%s", token.Name, token.Id)
			if token.Status != "" {
				line += "\t" + token.Status
			}
			if token.ExpiresAt != "" {
				line += "\texpires " + token.ExpiresAt
			}
			fmt.Println(line)
		}
	},
}

var createTokenCmd = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(1)
	},
}

var createAdminTokenCmd = &cobra.Command{
	Use:   "admin <token-name>",
	Args:  cobra.ExactArgs(1),
	Short: "Create an admin token",
	Long:  "You can use CDK_USER and CDK_PASSWORD environment variables to authenticate instead of CDK_API_KEY, in order to create your first token",
	Run: func(cmd *cobra.Command, args []string) {
		username := os.Getenv("CDK_USER")
		password := os.Getenv("CDK_PASSWORD")

		if username != "" && password == "" {
			fmt.Fprintln(os.Stderr, "Please set CDK_PASSWORD if you set CDK_USER")
			os.Exit(2)
		} else if username == "" && password != "" {
			fmt.Fprintln(os.Stderr, "Please set CDK_USER if you set CDK_PASSWORD")
			os.Exit(3)
		} else if username != "" && password != "" {
			jwtToken, err := consoleAPIClient().Login(username, password)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not login: %s\n", err)
				os.Exit(4)
			}
			consoleAPIClient().SetAPIKey(jwtToken.AccessToken)
		} else if os.Getenv("CDK_API_KEY") == "" {
			fmt.Fprintln(os.Stderr, "Please set CDK_API_KEY or CDK_USER and CDK_PASSWORD")
			os.Exit(5)
		}

		validateExpiresAt(*expiresAtForCreateAdmin)
		result, err := consoleAPIClient().CreateAdminToken(args[0], *expiresAtForCreateAdmin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create admin token: %s\n", err)
			os.Exit(4)
		}
		fmt.Printf("%s\n", result.Token)
	},
}

var createApplicationInstanceTokenCmd = &cobra.Command{
	Use:  "application-instance --application-instance=myappinstance <token-name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := consoleAPIClient().CreateApplicationInstanceToken(*applicationInstanceNameForCreate, args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create application-instance token: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", result.Token)
	},
}

var createApplicationTokenCmd = &cobra.Command{
	Use:  "application --application=myapp <token-name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateExpiresAt(*expiresAtForCreateApplication)
		result, err := consoleAPIClient().CreateApplicationToken(*applicationNameForCreate, args[0], *expiresAtForCreateApplication)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create application token: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", result.Token)
	},
}

var deleteTokenCmd = &cobra.Command{
	Use:  "delete <token-uuid>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := consoleAPIClient().DeleteToken(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not delete token: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Token deleted")
	},
}

var revokeTokenCmd = &cobra.Command{
	Use:   "revoke <token-uuid>",
	Args:  cobra.ExactArgs(1),
	Short: "Revoke a token",
	Long:  "Revoke a token so it can no longer be used. Unlike delete, revocation is terminal: the token stays listed with a Revoked status",
	Run: func(cmd *cobra.Command, args []string) {
		err := consoleAPIClient().RevokeToken(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not revoke token: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Token revoked")
	},
}

func init() {
	applicationInstanceNameForList = listApplicationInstanceTokenCmd.PersistentFlags().StringP("application-instance", "i", "", "Application instance name")
	applicationInstanceNameForCreate = createApplicationInstanceTokenCmd.PersistentFlags().StringP("application-instance", "i", "", "Application instance name")
	_ = listApplicationInstanceTokenCmd.MarkPersistentFlagRequired("application-instance")
	_ = createApplicationInstanceTokenCmd.MarkPersistentFlagRequired("application-instance")
	applicationNameForList = listApplicationTokenCmd.PersistentFlags().StringP("application", "a", "", "Application name")
	applicationNameForCreate = createApplicationTokenCmd.PersistentFlags().StringP("application", "a", "", "Application name")
	_ = listApplicationTokenCmd.MarkPersistentFlagRequired("application")
	_ = createApplicationTokenCmd.MarkPersistentFlagRequired("application")
	expiresAtForCreateAdmin = createAdminTokenCmd.PersistentFlags().String("expires-at", "", "Token expiration date (RFC3339, e.g. 2027-01-01T00:00:00Z); never expires if not set")
	expiresAtForCreateApplication = createApplicationTokenCmd.PersistentFlags().String("expires-at", "", "Token expiration date (RFC3339, e.g. 2027-01-01T00:00:00Z); never expires if not set")
	listTokenCmd.AddCommand(listApplicationInstanceTokenCmd)
	listTokenCmd.AddCommand(listApplicationTokenCmd)
	listTokenCmd.AddCommand(listAdminCmd)
	tokenCmd.AddCommand(listTokenCmd)
	tokenCmd.AddCommand(deleteTokenCmd)
	tokenCmd.AddCommand(revokeTokenCmd)
	tokenCmd.AddCommand(createTokenCmd)
	createTokenCmd.AddCommand(createAdminTokenCmd)
	createTokenCmd.AddCommand(createApplicationInstanceTokenCmd)
	createTokenCmd.AddCommand(createApplicationTokenCmd)
	rootCmd.AddCommand(tokenCmd)
}
