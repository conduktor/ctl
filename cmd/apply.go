package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/client"
	"github.com/conduktor/ctl/resource"
	"github.com/spf13/cobra"
	"os"
)

var filePath *string

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "upsert a resource on Conduktor",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var resources []resource.Resource
		var err error

		directory, err := isDirectory(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if directory {
			resources, err = resource.FromFolder(*filePath)
		} else {
			resources, err = resource.FromFile(*filePath)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		client := client.MakeFromEnv(*debug)
		for _, resource := range resources {
			upsertResult, err := client.Apply(&resource)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not apply resource %s/%s: %s\n", resource.Kind, resource.Name, err)
				os.Exit(1)
			} else {
				fmt.Printf("%s/%s: %s\n", resource.Kind, resource.Name, upsertResult)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.
	filePath = applyCmd.
		PersistentFlags().StringP("file", "f", "", "Specify the file to apply")

	applyCmd.MarkPersistentFlagRequired("file")
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
