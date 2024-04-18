package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

func initPrintKind(kinds schema.KindCatalog) {

	var prettyPrint *bool

	var printKind = &cobra.Command{
		Use:    "printKind",
		Short:  "Print kind catalog used",
		Long:   ``,
		Args:   cobra.NoArgs,
		Hidden: true,
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
