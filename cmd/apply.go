package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/client"
	"github.com/conduktor/ctl/resource"
	"github.com/spf13/cobra"
	"os"
)

var filePath *[]string

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "upsert a resource on Conduktor",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var resources []resource.Resource = make([]resource.Resource, 0)
		for _, path := range *filePath {
			r, err := resourceForPath(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			resources = append(resources, r...)
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

func resourceForPath(path string) ([]resource.Resource, error) {
	directory, err := isDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if directory {
		return resource.FromFolder(path)
	} else {
		return resource.FromFile(path)
	}
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.
	filePath = applyCmd.
		PersistentFlags().StringArrayP("file", "f", make([]string, 0, 0), "Specify the files to apply")

	applyCmd.MarkPersistentFlagRequired("file")
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
