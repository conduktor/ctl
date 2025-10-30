package cmd

import (
	"github.com/conduktor/ctl/internal/schema"
	"github.com/conduktor/ctl/internal/utils"
	"github.com/spf13/cobra"
)

func intConsoleMakeCatalog() {
	var prettyPrint *bool
	var nonStrict *bool

	var makeKind = &cobra.Command{
		Use:     "makeConsoleCatalog [file]",
		Short:   "Make catalog json from openapi file if file not given it will read from api",
		Long:    ``,
		Aliases: []string{"mkKind", "makeConsoleKind", "makeKind"}, // for backward compatibility
		Args:    cobra.RangeArgs(0, 1),
		Hidden:  !utils.CdkDevMode(),
		Run: func(cmd *cobra.Command, args []string) {
			runMakeCatalog(cmd, args, *prettyPrint, *nonStrict, func() ([]byte, error) { return consoleAPIClient().GetOpenAPI() }, func(schema *schema.OpenAPIParser, strict bool) (*schema.Catalog, error) {
				return schema.GetConsoleCatalog(strict)
			})
		},
	}
	rootCmd.AddCommand(makeKind)

	prettyPrint = makeKind.Flags().BoolP("pretty", "p", false, "Pretty print the output")
	nonStrict = makeKind.Flags().BoolP("non-strict", "n", false, "Don't be strict on the parsing of the schema")
}
