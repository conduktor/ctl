package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

func runMkKind(cmd *cobra.Command, args []string, prettyPrint bool, nonStrict bool, getOpenApi func() ([]byte, error), getKinds func(*schema.OpenApiParser, bool) (schema.KindCatalog, error)) {
	var kinds map[string]schema.Kind
	if len(args) == 1 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			panic(err)
		}
		schema, err := schema.NewOpenApiParser(data)
		if err != nil {
			panic(err)
		}
		kinds, err = getKinds(schema, !nonStrict)
		if err != nil {
			panic(err)
		}
	} else {
		data, err := getOpenApi()
		if err != nil {
			panic(err)
		}
		schema, err := schema.NewOpenApiParser(data)
		if err != nil {
			panic(err)
		}
		kinds, err = getKinds(schema, !nonStrict)
		if err != nil {
			panic(err)
		}
	}
	var payload []byte
	var err error
	if prettyPrint {
		payload, err = json.MarshalIndent(kinds, "", "  ")
	} else {
		payload, err = json.Marshal(kinds)
	}
	if err != nil {
		panic(err)
	}
	fmt.Print(string(payload))
}
