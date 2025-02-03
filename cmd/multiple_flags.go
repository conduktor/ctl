package cmd

import (
	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/spf13/cobra"
	"strconv"
)

type MultipleFlags struct {
	result     map[string]interface{}
	flagParams map[string]schema.FlagParameterOption
	command    *cobra.Command
}

func NewMultipleFlags(command *cobra.Command, flagParams map[string]schema.FlagParameterOption) *MultipleFlags {
	result := make(map[string]interface{}, len(flagParams))
	usage := ""
	for key, flag := range flagParams {
		var isFlagSet bool
		if flag.Type == "string" {
			isFlagSet = true
			defaultValue := ""
			result[key] = command.Flags().String(flag.FlagName, defaultValue, usage)
		} else if flag.Type == "boolean" {
			isFlagSet = true
			defaultValue := false
			result[key] = command.Flags().Bool(flag.FlagName, defaultValue, usage)
		} else if flag.Type == "integer" {
			isFlagSet = true
			defaultValue := 0
			result[key] = command.Flags().Int(flag.FlagName, defaultValue, usage)
		} else if utils.CdkDebug() {
			println("Unknown flag type: " + flag.Type)
		}
		if isFlagSet && flag.Required {
			command.MarkFlagRequired(flag.FlagName)
		}
	}

	return &MultipleFlags{
		result,
		flagParams,
		command,
	}
}

func (m *MultipleFlags) ExtractFlagValueForBodyParam() map[string]interface{} {
	bodyParams := make(map[string]interface{})
	for key, value := range m.result {
		if value != nil && m.flagSetByUser(key) {
			bodyParams[key] = value
		}
	}
	return bodyParams
}

func (m *MultipleFlags) ExtractFlagValueForQueryParam() map[string]string {
	queryParams := make(map[string]string)
	for key, value := range m.result {
		if value != nil && m.flagSetByUser(key) {
			str, strOk := value.(*string)
			boolValue, boolOk := value.(*bool)
			intValue, intOk := value.(*int)

			if strOk {
				queryParams[key] = *str
			} else if boolOk {
				queryParams[key] = strconv.FormatBool(*boolValue)
			} else if intOk {
				queryParams[key] = strconv.Itoa(*intValue)
			} else {
				panic("Unknown query flag type")
			}
		}
	}
	return queryParams
}

func (m *MultipleFlags) flagSetByUser(flagKey string) bool {
	flag, present := m.flagParams[flagKey]
	if !present {
		panic("Flag " + flagKey + " not found in flagParams")
	}
	return m.command.Flags().Changed(flag.FlagName)
}
