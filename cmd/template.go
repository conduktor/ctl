package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

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

func initTemplate(kinds schema.KindCatalog, strict bool) {
	rootCmd.AddCommand(templateCmd)
	var file *string
	var edit *bool
	var apply *bool
	file = templateCmd.PersistentFlags().StringP("output", "o", "", "Write example to file")
	edit = templateCmd.PersistentFlags().BoolP("edit", "e", false, "Edit the file after it's creation")
	apply = templateCmd.PersistentFlags().BoolP("apply", "a", false, "Apply the yaml file after it's edition")

	// Add all kinds to the 'template' command
	for name, kind := range kinds {
		kindCmd := &cobra.Command{
			Use:     name,
			Short:   "Get a yaml example for resource of kind " + name,
			Args:    cobra.NoArgs,
			Long:    `If name not provided it will list all resource`,
			Aliases: buildAlias(name),
			PreRun: func(cmd *cobra.Command, args []string) {
				if edit != nil && *edit && (file == nil || *file == "") {
					fmt.Fprintln(os.Stderr, "Cannot use --edit without --output")
					os.Exit(10)
				}
				if apply != nil && *apply && (edit == nil || !*edit) {
					fmt.Fprintln(os.Stderr, "Cannot use --apply without --edit")
					os.Exit(11)
				}
			},
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
						if apply != nil && *apply {
							_, err = w.WriteString("WARNING: Your file will be applied automatically once saved. If you do not want to apply anything, save an empty file.\n")
						}
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
						editAndApply(edit, file, apply, kinds, strict)
					}
				}
			},
		}
		templateCmd.AddCommand(kindCmd)
	}
}

func editAndApply(edit *bool, file *string, apply *bool, kinds schema.KindCatalog, strict bool) {
	if edit != nil && *edit {
		//run $EDITOR on the file
		editor := os.Getenv("EDITOR")
		if editor == "" {
			fmt.Fprintln(os.Stderr, "No editor set. Set $EDITOR to your preferred editor")
			os.Exit(6)
		}
		editorFromPath, err := exec.LookPath(editor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not find $EDITOR %s in path: %s\n", editor, err)
			os.Exit(7)
		}
		cmd := exec.Command(editorFromPath, *file)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Could not run %s: %s", editorFromPath, err)
			os.Exit(8)
		}
		if apply != nil && *apply {
			runApply(kinds, []string{*file}, strict)
		}
	}
}
