package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
)

type OutputFormat enumflag.Flag

const (
	JSON OutputFormat = iota
	YAML
	NAME
)

var OutputFormatIds = map[OutputFormat][]string{
	JSON: {"json"},
	YAML: {"yaml"},
	NAME: {"name"},
}

func (o OutputFormat) String() string {
	return OutputFormatIds[o][0]
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resource of a given kind",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		_ = cmd.Help()
		os.Exit(1)
	},
}

func initGet(rootContext cli.RootContext) {
	var format OutputFormat = YAML
	getCmd.PersistentFlags().VarP(enumflag.New(&format, "output", OutputFormatIds, enumflag.EnumCaseInsensitive), "output", "o", "Output format. One of: json|yaml|name")
	rootCmd.AddCommand(getCmd)

	var onlyGateway *bool
	var onlyConsole *bool
	var allCmd = &cobra.Command{
		Use:   "all",
		Short: "Get all global resources",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			getAllCommandRun(rootContext, onlyGateway, onlyConsole, format)
		},
	}
	onlyGateway = allCmd.Flags().BoolP("gateway", "g", false, "Only show gateway resources")
	onlyConsole = allCmd.Flags().BoolP("console", "c", false, "Only show console resources")
	allCmd.MarkFlagsMutuallyExclusive("gateway", "console")
	getCmd.AddCommand(allCmd)

	// Add all kinds to the 'get' command
	for name, kind := range rootContext.Catalog.Kind {
		gatewayKind, isGatewayKind := kind.GetLatestKindVersion().(*schema.GatewayKindVersion)
		args := cobra.MaximumNArgs(1)
		use := fmt.Sprintf("%s [name]", name)
		if isGatewayKind && !gatewayKind.GetAvailable {
			args = cobra.NoArgs
			use = name
		}
		parentFlags := kind.GetParentFlag()
		parentQueryFlags := kind.GetParentQueryFlag()
		parentFlagValue := make([]*string, len(parentFlags))
		parentQueryFlagValue := make([]*string, len(parentQueryFlags))
		var multipleFlags *MultipleFlags
		kindCmd := &cobra.Command{
			Use:     use,
			Short:   "Get resource of kind " + name,
			Args:    args,
			Long:    `If name not provided it will list all resource`,
			Aliases: buildAlias(name),
			Run: func(cmd *cobra.Command, args []string) {
				getKindCommandRun(rootContext, kind, args, parentFlagValue, parentQueryFlagValue, multipleFlags, format)
			},
		}
		for i, flag := range parentFlags {
			parentFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
			_ = kindCmd.MarkFlagRequired(flag)
		}
		for i, flag := range parentQueryFlags {
			parentQueryFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
		}
		multipleFlags = NewMultipleFlags(kindCmd, kind.GetListFlag())
		getCmd.AddCommand(kindCmd)
	}
}

func getAllCommandRun(rootContext cli.RootContext, onlyGateway *bool, onlyConsole *bool, format OutputFormat) {
	cmdCtx := cli.GetAllHandlerContext{
		OnlyGateway: onlyGateway,
		OnlyConsole: onlyConsole,
	}

	allResources, errors := cli.GetAllsHandler(rootContext, cmdCtx)
	for _, err := range errors {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	err := printResource(allResources, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func getKindCommandRun(
	rootContext cli.RootContext,
	kind schema.Kind,
	args []string,
	parentFlagValue []*string,
	parentQueryFlagValue []*string,
	multipleFlags *MultipleFlags,
	format OutputFormat) {

	cmdCtx := cli.GetKindHandlerContext{
		Args:                 args,
		ParentFlagValue:      parentFlagValue,
		ParentQueryFlagValue: parentQueryFlagValue,
		QueryParams:          multipleFlags.ExtractFlagValueForQueryParam(),
	}

	result, errors := cli.GetKindHandler(kind, rootContext, cmdCtx)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(1)
	}

	err := printResource(result, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func printResource(result interface{}, format OutputFormat) error {
	switch format {
	case JSON:
		jsonOutput, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON: %s\n%s", err, result)
		}
		fmt.Println(string(jsonOutput))
	case NAME:
		// show Kind/Name
		switch res := result.(type) {
		case []resource.Resource:
			for _, r := range res {
				fmt.Println(r.Kind + "/" + r.Name)
			}
		case resource.Resource:
			fmt.Println(res.Kind + "/" + res.Name)
		default:
			return fmt.Errorf("unexpected resource type")
		}
	case YAML:
		switch res := result.(type) {
		case []resource.Resource:
			for _, r := range res {
				fmt.Println("---") // '---' indicates the start of a new document in YAML
				_ = r.PrintPreservingOriginalFieldOrder()
			}
		case resource.Resource:
			_ = res.PrintPreservingOriginalFieldOrder()
		default:
			return fmt.Errorf("unexpected resource type")
		}
	default:
		return fmt.Errorf("invalid output format %s", format.String())
	}
	return nil
}
