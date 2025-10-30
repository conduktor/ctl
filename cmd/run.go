package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/schema"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run an action",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		_ = cmd.Help()
		os.Exit(1)
	},
}

func printExecuteResult(result []byte) {
	var stringJSON string
	err := json.Unmarshal(result, &stringJSON)
	if err == nil {
		fmt.Println(stringJSON)
		return
	}
	var jsonResult interface{}
	err = json.Unmarshal(result, &jsonResult)
	if err == nil {
		jsonOutput, err := json.MarshalIndent(jsonResult, "", "  ")
		if err == nil {
			fmt.Println(string(jsonOutput))
			return
		}
	}
	if len(result) > 1 {
		fmt.Println(string(result))
	} else {
		fmt.Println("Ok")
	}
}

func initRun(runs schema.RunCatalog) {
	rootCmd.AddCommand(runCmd)

	for name, run := range runs {
		args := cobra.MaximumNArgs(0)
		pathFlags := run.PathParameter
		pathFlagValues := make([]*string, len(pathFlags))
		queryFlags := run.QueryParameter
		bodyFlags := run.BodyFields
		var multipleFlagsForQuery *MultipleFlags
		var multipleFlagsForBody *MultipleFlags
		subRunCmd := &cobra.Command{
			Use:     name,
			Short:   run.Doc,
			Args:    args,
			Aliases: buildAlias(name),
			Run: func(cmd *cobra.Command, args []string) {
				pathValues := make([]string, len(pathFlagValues))
				queryParams := multipleFlagsForQuery.ExtractFlagValueForQueryParam()
				body := multipleFlagsForBody.ExtractFlagValueForBodyParam()
				for i, v := range pathFlagValues {
					pathValues[i] = *v
				}

				var err error

				if len(bodyFlags) == 0 {
					body = nil
				}
				var result []byte
				switch run.BackendType {
				case schema.CONSOLE:
					result, err = consoleAPIClient().Run(run, pathValues, queryParams, body)
				case schema.GATEWAY:
					result, err = gatewayAPIClient().Run(run, pathValues, queryParams, body)
				default:
					panic("Unknown backend type")
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fetching resources: %s\n", err)
					return
				}
				printExecuteResult(result)
			},
		}
		for i, flag := range pathFlags {
			pathFlagValues[i] = subRunCmd.Flags().String(flag, "", "Parent "+flag)
			err := subRunCmd.MarkFlagRequired(flag)
			if err != nil {
				panic(err)
			}
		}
		multipleFlagsForQuery = NewMultipleFlags(subRunCmd, queryFlags)
		multipleFlagsForBody = NewMultipleFlags(subRunCmd, bodyFlags)

		runCmd.AddCommand(subRunCmd)
	}
}
