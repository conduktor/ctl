package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resource of a given kind",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		cmd.Help()
		os.Exit(1)
	},
}

func buildQueryParams(params map[string]interface{}) map[string]string {
	queryParams := make(map[string]string)
	for key, value := range params {
		if value != nil {
			str, strOk := value.(*string)
			boolValue, boolOk := value.(*bool)

			if strOk {
				if *str != "" {
					queryParams[key] = *str
				}
			} else if boolOk {
				queryParams[key] = strconv.FormatBool(*boolValue)
			} else {
				panic("Unknown query flag type")
			}
		}
	}
	return queryParams
}

func removeTrailingSIfAny(name string) string {
	return strings.TrimSuffix(name, "s")
}

func buildAlias(name string) []string {
	return []string{strings.ToLower(name), removeTrailingSIfAny(strings.ToLower(name)), removeTrailingSIfAny(name)}
}

func initGet(kinds schema.KindCatalog) {
	rootCmd.AddCommand(getCmd)

	for name, kind := range kinds {
		gatewayKind, isGatewayKind := kind.GetLatestKindVersion().(*schema.GatewayKindVersion)
		args := cobra.MaximumNArgs(1)
		use := fmt.Sprintf("%s [name]", name)
		if isGatewayKind && !gatewayKind.GetAvailable {
			args = cobra.NoArgs
			use = fmt.Sprintf("%s", name)
		}
		parentFlags := kind.GetParentFlag()
		listFlags := kind.GetListFlag()
		parentFlagValue := make([]*string, len(parentFlags))
		listFlagValue := make(map[string]interface{}, len(listFlags))
		kindCmd := &cobra.Command{
			Use:     use,
			Short:   "Get resource of kind " + name,
			Args:    args,
			Long:    `If name not provided it will list all resource`,
			Aliases: buildAlias(name),
			Run: func(cmd *cobra.Command, args []string) {
				parentValue := make([]string, len(parentFlagValue))
				queryParams := buildQueryParams(listFlagValue)
				for i, v := range parentFlagValue {
					parentValue[i] = *v
				}
				var err error
				if len(args) == 0 {
					var result []resource.Resource
					if isGatewayKind {
						result, err = gatewayApiClient().Get(&kind, parentValue, queryParams)
					} else {
						result, err = consoleApiClient().Get(&kind, parentValue, queryParams)
					}
					for _, r := range result {
						r.PrintPreservingOriginalFieldOrder()
						fmt.Println("---")
					}
				} else if len(args) == 1 {
					var result resource.Resource
					if isGatewayKind {
						result, err = gatewayApiClient().Describe(&kind, parentValue, args[0])
					} else {
						result, err = consoleApiClient().Describe(&kind, parentValue, args[0])
					}
					result.PrintPreservingOriginalFieldOrder()
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			},
		}
		for i, flag := range parentFlags {
			parentFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
			kindCmd.MarkFlagRequired(flag)
		}
		for key, flag := range listFlags {
			var flagSetted = false
			if flag.Type == "string" {
				flagSetted = true
				listFlagValue[key] = kindCmd.Flags().String(flag.FlagName, "", "")
			} else if flag.Type == "boolean" {
				flagSetted = true
				listFlagValue[key] = kindCmd.Flags().Bool(flag.FlagName, false, "")
			} else {
				if *debug || utils.CdkDebug() {
					fmt.Fprintf(os.Stderr, "Unknown flag type %s\n", flag.Type)
				}
			}
			if flagSetted && flag.Required {
				kindCmd.MarkFlagRequired(flag.FlagName)
			}
		}
		getCmd.AddCommand(kindCmd)
	}
}
