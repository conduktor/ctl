package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/conduktor/ctl/internal/resource"
	"github.com/conduktor/ctl/internal/schema"
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

func removeTrailingSIfAny(name string) string {
	return strings.TrimSuffix(name, "s")
}

func buildAlias(name string) []string {
	aliases := []string{strings.ToLower(name)}
	// This doesn't seem to be needed since none of the kinds ends with an S
	// However I'm leaving it here as a conditional so it won't affect the usage
	if strings.HasSuffix(name, "s") {
		aliases = append(aliases, removeTrailingSIfAny(name), removeTrailingSIfAny(strings.ToLower(name)))
	}
	return aliases
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

func isGateway(kind schema.Kind) bool {
	_, isGatewayKind := kind.GetLatestKindVersion().(*schema.GatewayKindVersion)
	return isGatewayKind
}

func isConsole(kind schema.Kind) bool {
	_, isConsoleKind := kind.GetLatestKindVersion().(*schema.ConsoleKindVersion)
	return isConsoleKind
}

func initGet(kinds schema.KindCatalog) {
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
			var allResources []resource.Resource

			kindsByName := sortedKeys(kinds)
			if gatewayAPIClientError != nil {
				if *debug || *onlyGateway {
					fmt.Fprintf(os.Stderr, "Cannot create Gateway client: %s\n", gatewayAPIClientError)
				}
			}
			if consoleAPIClientError != nil {
				if *debug || *onlyConsole {
					fmt.Fprintf(os.Stderr, "Cannot create Console client: %s\n", consoleAPIClientError)

				}
			}
			for _, key := range kindsByName {
				kind := kinds[key]
				// keep only the Kinds where listing is provided TODO fix if config is provided
				if !kind.IsRootKind() {
					continue
				}
				var resources []resource.Resource
				var err error
				if isGateway(kind) && !*onlyConsole && gatewayAPIClientError == nil {
					resources, err = gatewayAPIClient().Get(&kind, []string{}, []string{}, map[string]string{})
				} else if isConsole(kind) && !*onlyGateway && consoleAPIClientError == nil {
					resources, err = consoleAPIClient().Get(&kind, []string{}, []string{}, map[string]string{})
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fetching resource %s: %s\n", kind.GetName(), err)
					continue
				}

				allResources = append(allResources, resources...)
			}
			err := printResource(allResources, format)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}
	onlyGateway = allCmd.Flags().BoolP("gateway", "g", false, "Only show gateway resources")
	onlyConsole = allCmd.Flags().BoolP("console", "c", false, "Only show console resources")
	allCmd.MarkFlagsMutuallyExclusive("gateway", "console")
	getCmd.AddCommand(allCmd)

	// Add all kinds to the 'get' command
	for name, kind := range kinds {
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
				parentValue := make([]string, len(parentFlagValue))
				parentQueryValue := make([]string, len(parentQueryFlagValue))
				queryParams := multipleFlags.ExtractFlagValueForQueryParam()
				for i, v := range parentFlagValue {
					parentValue[i] = *v
				}
				for i, v := range parentQueryFlagValue {
					parentQueryValue[i] = *v
				}

				var err error

				if len(args) == 0 {
					var result []resource.Resource
					if isGatewayKind {
						result, err = gatewayAPIClient().Get(&kind, parentValue, parentQueryValue, queryParams)
					} else {
						result, err = consoleAPIClient().Get(&kind, parentValue, parentQueryValue, queryParams)
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error fetching resources: %s\n", err)
						return
					}
					err = printResource(result, format)
				} else if len(args) == 1 {
					var result resource.Resource
					if isGatewayKind {
						result, err = gatewayAPIClient().Describe(&kind, parentValue, parentQueryValue, args[0])
					} else {
						result, err = consoleAPIClient().Describe(&kind, parentValue, parentQueryValue, args[0])
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error describing resource: %s\n", err)
						return
					}
					err = printResource(result, format)
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
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

func sortedKeys(kinds schema.KindCatalog) []string {
	keys := make([]string, 0, len(kinds))
	for key := range kinds {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
