package cmd

import (
	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/spf13/cobra"
)

func initGatewayMkKind() {
	var prettyPrint *bool
	var nonStrict *bool

	var makeKind = &cobra.Command{
		Use:     "gatewayMakeKind [file]",
		Short:   "Make kind json from openapi file if file not given it will read from api",
		Long:    ``,
		Aliases: []string{"gatewayMkKind", "gwMkKind"},
		Args:    cobra.RangeArgs(0, 1),
		Hidden:  !utils.CdkDevMode(),
		Run: func(cmd *cobra.Command, args []string) {
			runMkKind(cmd, args, *prettyPrint, *nonStrict, func() ([]byte, error) { return gatewayApiClient().GetOpenApi() }, func(schema *schema.Schema, strict bool) (schema.KindCatalog, error) {
				return schema.GetGatewayKinds(strict)
			})
		},
	}
	rootCmd.AddCommand(makeKind)

	prettyPrint = makeKind.Flags().BoolP("pretty", "p", false, "Pretty print the output")
	nonStrict = makeKind.Flags().BoolP("non-strict", "n", false, "Don't be strict on the parsing of the schema")
}
