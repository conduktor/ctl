package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Get a yaml example for a given kind",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		cmd.Help()
		os.Exit(1)
	},
}

func initTemplate(kinds schema.KindCatalog) {
	rootCmd.AddCommand(templateCmd)
	var file *string
	file = templateCmd.PersistentFlags().StringP("output", "o", "", "Write example to file")

	// Add all kinds to the 'template' command
	for name, kind := range kinds {
		kindCmd := &cobra.Command{
			Use:     name,
			Short:   "Get a yaml example for resource of kind " + name,
			Args:    cobra.NoArgs,
			Long:    `If name not provided it will list all resource`,
			Aliases: buildAlias(name),
			Run: func(cmd *cobra.Command, args []string) {
				example := kind.GetLatestKindVersion().GetApplyExample()
				if example == "" {
					fmt.Fprintf(os.Stderr, "No template for kind %s\n", name)
					os.Exit(1)
				} else {
					if file == nil || *file == "" {
						fmt.Println("---")
						fmt.Println(kind.GetLatestKindVersion().GetApplyExample())
					} else {
						_, err := os.Stat(*file)
						if err == nil {
							fmt.Fprintf(os.Stderr, "File %s already exists. You can use conduktor template %s >> %s to append to existing file\n", *file, name, *file)
							os.Exit(2)
						}
						f, err := os.Create(*file)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error creating file %s: %s\n", *file, err)
							os.Exit(3)
						}
						defer f.Close()
						w := bufio.NewWriter(f)
						_, err = w.WriteString("---\n")
						_, err = w.WriteString(kind.GetLatestKindVersion().GetApplyExample())
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error writting to file %s: %s\n", *file, err)
							os.Exit(4)
						}
						err = w.Flush()
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error writting to file %s: %s\n", *file, err)
							os.Exit(5)
						}
					}
				}
			},
		}
		templateCmd.AddCommand(kindCmd)
	}
}
