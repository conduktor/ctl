package cmd

import (
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
	"strconv"
)

// used to detect when flag wasn't used
const unprobableInt = -23871

// for string it's ""
// for bool we will keep default false value

func extractFlagValueForQueryParam(params map[string]interface{}) map[string]string {
	queryParams := make(map[string]string)
	for key, value := range params {
		if value != nil {
			str, strOk := value.(*string)
			boolValue, boolOk := value.(*bool)
			intValue, intOk := value.(*int)

			if strOk {
				if *str != "" {
					queryParams[key] = *str
				}
			} else if boolOk {
				queryParams[key] = strconv.FormatBool(*boolValue)
			} else if intOk {
				if *intValue != unprobableInt {
					queryParams[key] = strconv.Itoa(*intValue)
				}
			} else {
				panic("Unknown query flag type")
			}
		}
	}
	return queryParams
}

func extractFlagValueForBodyParam(params map[string]interface{}) map[string]interface{} {
	bodyParams := make(map[string]interface{})
	for key, value := range params {
		if value != nil {
			str, strOk := value.(*string)
			boolValue, boolOk := value.(*bool)
			intValue, intOk := value.(*int)

			if strOk {
				if *str != "" {
					bodyParams[key] = str
				}
			} else if boolOk {
				bodyParams[key] = boolValue
			} else if intOk {
				if *intValue != unprobableInt {
					bodyParams[key] = *intValue
				}
			} else {
				bodyParams[key] = value
			}
		}
	}
	return bodyParams
}

func buildFlag(command *cobra.Command, flagParams map[string]schema.FlagParameterOption, flagResult map[string]interface{}) {
	for key, flag := range flagParams {
		var isFlagSet bool
		if flag.Type == "string" {
			isFlagSet = true
			flagResult[key] = command.Flags().String(flag.FlagName, "", "")
		} else if flag.Type == "boolean" {
			isFlagSet = true
			flagResult[key] = command.Flags().Bool(flag.FlagName, false, "")
		} else if flag.Type == "integer" {
			isFlagSet = true
			flagResult[key] = command.Flags().Int(flag.FlagName, unprobableInt, "")
		}
		if isFlagSet && flag.Required {
			command.MarkFlagRequired(flag.FlagName)
		}
	}
}
