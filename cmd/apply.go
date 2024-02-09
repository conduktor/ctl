/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/resource"
	"github.com/spf13/cobra"
	"os"
)

var filePath *string

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "upsert a resource on kubernetes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		resources, error := resource.FromFile(*filePath)
		if error != nil {
			fmt.Fprintf(os.Stderr, "%s\n", error)
			os.Exit(1)
		}
		fmt.Println(resources)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.
	filePath = applyCmd.
		PersistentFlags().StringP("file", "f", "", "Specify the file to apply")

	applyCmd.MarkPersistentFlagRequired("file")
}
