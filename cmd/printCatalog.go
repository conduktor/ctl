package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/spf13/cobra"
)

func initPrintCatalog(kinds schema.Catalog) {

	var prettyPrint *bool

	var printKind = &cobra.Command{
		Use:     "printCatalog",
		Aliases: []string{"printKind"}, // for backward compatibility
		Short:   "Print catalog",
		Long:    ``,
		Args:    cobra.NoArgs,
		Hidden:  !utils.CdkDevMode(),
		Run: func(cmd *cobra.Command, args []string) {
			var payload []byte
			var err error
			if *prettyPrint {
				payload, err = json.MarshalIndent(kinds, "", "  ")
			} else {
				payload, err = json.Marshal(kinds)
			}
			if err != nil {
				panic(err)
			}
			fmt.Print(string(payload))
		},
	}

	rootCmd.AddCommand(printKind)

	prettyPrint = printKind.Flags().BoolP("pretty", "p", false, "Pretty print the output")
}
