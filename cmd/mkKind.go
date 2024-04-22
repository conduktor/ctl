package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
	"os"
)

func initMkKind() {
	var prettyPrint *bool
	var nonStrict *bool

	var makeKind = &cobra.Command{
		Use:    "makeKind [file]",
		Short:  "Make kind json from openapi file if file not given it will read from api",
		Long:   ``,
		Args:   cobra.RangeArgs(0, 1),
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			var kinds map[string]schema.Kind
			if len(args) == 1 {
				data, err := os.ReadFile(args[0])
				if err != nil {
					panic(err)
				}
				schema, err := schema.New(data)
				if err != nil {
					panic(err)
				}
				kinds, err = schema.GetKinds(!*nonStrict)
				if err != nil {
					panic(err)
				}
			} else {
				data, err := apiClient().GetOpenApi()
				if err != nil {
					panic(err)
				}
				schema, err := schema.New(data)
				if err != nil {
					panic(err)
				}
				kinds, err = schema.GetKinds(!*nonStrict)
				if err != nil {
					panic(err)
				}
			}
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
	rootCmd.AddCommand(makeKind)

	prettyPrint = makeKind.Flags().BoolP("pretty", "p", false, "Pretty print the output")
	nonStrict = makeKind.Flags().BoolP("non-strict", "n", false, "Don't be strict on the parsing of the schema")
}
