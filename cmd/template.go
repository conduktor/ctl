package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/conduktor/ctl/internal/printutils"
	"github.com/conduktor/ctl/internal/utils"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Get a yaml example for a given kind",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		_ = cmd.Help()
		os.Exit(1)
	},
}

func initTemplate(rootContext cli.RootContext) {
	rootCmd.AddCommand(templateCmd)
	var file *string
	var edit *bool
	var apply *bool
	file = templateCmd.PersistentFlags().StringP("output", "o", "", "Write example to file")
	edit = templateCmd.PersistentFlags().BoolP("edit", "e", false, "Edit the YAML file post-creation; this works only with --output. It will the EDITOR environment variable or nano if not set.")
	apply = templateCmd.PersistentFlags().BoolP("apply", "a", false, "Apply the YAML file post-editing; this works only with --edit.")

	// Add all kinds to the 'template' command
	for name, kind := range rootContext.Catalog.Kind {
		kindCmd := &cobra.Command{
			Use:     name + " [template-name]",
			Short:   "Get a yaml example for resource of kind " + name,
			Args:    cobra.MaximumNArgs(1),
			Long:    `Without a name, returns a built-in example. With a name, fetches the matching server-side template (requires a Conduktor Console that supports resource templates).`,
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
				var example string
				if len(args) == 1 {
					var err error
					example, err = fetchTemplateByName(rootContext, name, args[0])
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s\n", err)
						os.Exit(1)
					}
				} else {
					example = kind.GetLatestKindVersion().GetApplyExample()
				}
				if example == "" {
					fmt.Fprintf(os.Stderr, "No template for kind %s\n", name)
					os.Exit(1)
				} else {
					if file == nil || *file == "" {
						fmt.Println("---")
						fmt.Println(example)
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
							_, err = w.WriteString(AutoApplyWarningMessage)
							if err != nil {
								fmt.Fprintf(os.Stderr, "Error writing to file %s: %s\n", *file, err)
								os.Exit(4)
							}
						}
						_, err = w.WriteString("---\n")
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error writing to file %s: %s\n", *file, err)
							os.Exit(4)
						}
						_, err = w.WriteString(example)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error writing to file %s: %s\n", *file, err)
							os.Exit(4)
						}
						err = w.Flush()
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error writing to file %s: %s\n", *file, err)
							os.Exit(5)
						}
						editAndApply(rootContext, edit, file, apply)
					}
				}
			},
		}
		templateCmd.AddCommand(kindCmd)
	}
}

// kindsSupportingTemplates lists the resource kinds Console exposes through the
// server-side template API. Other kinds don't have a `*-template` endpoint, so
// we fail fast instead of hitting the API and producing a confusing parse error.
var kindsSupportingTemplates = map[string]bool{
	"Topic":     true,
	"Connector": true,
}

// fetchTemplateByName fetches an admin-curated server-side template named
// `templateName` for the given resource kind (e.g. "Topic") and renders it as a YAML
// resource of that kind. The template's `spec.defaults` carries the metadata + spec
// that should be used when instantiating the underlying resource.
func fetchTemplateByName(rootContext cli.RootContext, kindName, templateName string) (string, error) {
	baseKind, ok := rootContext.Catalog.Kind[kindName]
	if !ok {
		return "", fmt.Errorf("Unknown kind %s", kindName)
	}

	if !kindsSupportingTemplates[kindName] {
		return "", fmt.Errorf("kind %s does not support resource templates (supported kinds: Topic, Connector)", kindName)
	}

	res, err := consoleAPIClient().GetTemplate(utils.CamelToKebab(kindName), templateName)
	if err != nil {
		return "", err
	}

	return printutils.RenderTemplateAsKind(res.Spec, baseKind.GetName(), baseKind.MaxVersion())
}

func editAndApply(rootContext cli.RootContext, edit *bool, file *string, apply *bool) {
	if edit != nil && *edit {
		// Run editor on the file
		err := runEditor(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Editor error: %s\n", err)
			os.Exit(7)
		}

		if apply != nil && *apply {
			filepath := []string{*file}
			cmdCtx := cli.ApplyHandlerContext{
				FilePaths:       filepath,
				RecursiveFolder: false,
				DryRun:          false,
				PrintDiff:       false,
				MaxParallel:     1,
				StateEnabled:    false,
				StateRef:        nil,
			}
			applyHandler := cli.NewApplyHandler(rootContext)

			_, err = applyHandler.Handle(cmdCtx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error during apply: %s\n", err)
				os.Exit(1)
			}
		}
	}
}
