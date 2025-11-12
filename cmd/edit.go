package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/conduktor/ctl/internal/orderedjson"
	"github.com/conduktor/ctl/internal/printutils"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a resource in a text editor and apply changes",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		_ = cmd.Help()
		os.Exit(1)
	},
}

func initEdit(rootContext cli.RootContext, kinds schema.KindCatalog) {
	rootCmd.AddCommand(editCmd)

	// Add all kinds to the 'edit' command
	for name, kind := range kinds {
		gatewayKind, isGatewayKind := kind.GetLatestKindVersion().(*schema.GatewayKindVersion)
		args := cobra.ExactArgs(1) // edit command requires a resource name
		use := fmt.Sprintf("%s <name>", name)

		// Skip kinds that don't support getting individual resources
		if isGatewayKind && !gatewayKind.GetAvailable {
			continue
		}

		parentFlags := kind.GetParentFlag()
		parentQueryFlags := kind.GetParentQueryFlag()
		parentFlagValue := make([]*string, len(parentFlags))
		parentQueryFlagValue := make([]*string, len(parentQueryFlags))

		kindCmd := &cobra.Command{
			Use:     use,
			Short:   "Edit resource of kind " + name,
			Args:    args,
			Long:    `Edit the specified resource in a text editor and apply changes after saving`,
			Aliases: buildAlias(name),
			Run: func(cmd *cobra.Command, args []string) {
				parentValue := make([]string, len(parentFlagValue))
				parentQueryValue := make([]string, len(parentQueryFlagValue))
				for i, v := range parentFlagValue {
					parentValue[i] = *v
				}
				for i, v := range parentQueryFlagValue {
					parentQueryValue[i] = *v
				}

				resourceName := args[0]

				// Get the resource
				var result resource.Resource
				var err error

				if isGatewayKind {
					result, err = gatewayAPIClient().Describe(&kind, parentValue, parentQueryValue, resourceName)
				} else {
					result, err = consoleAPIClient().Describe(&kind, parentValue, parentQueryValue, resourceName)
				}

				if err != nil {
					fmt.Fprintf(os.Stderr, "Error describing resource: %s\n", err)
					os.Exit(1)
				}

				// Create a temporary file with the resource YAML
				tmpFile, err := os.CreateTemp("", fmt.Sprintf("conduktor-edit-%s-%s-*.yaml", name, resourceName))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating temporary file: %s\n", err)
					os.Exit(1)
				}
				defer os.Remove(tmpFile.Name()) // Clean up temp file

				// Write the resource to the temp file
				w := bufio.NewWriter(tmpFile)
				_, err = w.WriteString(AutoApplyWarningMessage)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to temporary file: %s\n", err)
					os.Exit(1)
				}
				_, err = w.WriteString("---\n")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to temporary file: %s\n", err)
					os.Exit(1)
				}

				var data orderedjson.OrderedData
				err = json.Unmarshal(result.Json, &data)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error unmarshalling resource JSON: %s\n", err)
					os.Exit(1)
				}

				err = printutils.PrintResourceLikeYamlFile(w, data)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing resource YAML: %s\n", err)
					os.Exit(1)
				}

				err = w.Flush()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error flushing temporary file: %s\n", err)
					os.Exit(1)
				}
				tmpFile.Close()

				// Open the file in an editor
				err = runEditor(tmpFile.Name())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Editor error: %s\n", err)
					os.Exit(1)
				}

				content, err := os.ReadFile(tmpFile.Name())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading modified file: %s\n", err)
					os.Exit(1)
				}

				// If file is empty, don't apply
				if len(content) == 0 {
					fmt.Println("Empty file, no changes applied")
					return
				}

				recursiveFolder := false
				runApply(rootContext, []string{tmpFile.Name()}, recursiveFolder)
			},
		}

		// Add parent flags
		for i, flag := range parentFlags {
			parentFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
			_ = kindCmd.MarkFlagRequired(flag)
		}
		for i, flag := range parentQueryFlags {
			parentQueryFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
		}

		editCmd.AddCommand(kindCmd)
	}
}
