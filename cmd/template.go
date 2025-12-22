package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Get a yaml example for a given kind",
	Long: `Get a yaml example for a given kind.

By default, uses embedded templates from the CLI (works offline).

With --live flag, fetches templates directly from the Conduktor Console server.
This ensures templates are always up-to-date with the server's schema.

Examples:
  conduktor template                            # List available kinds (embedded)
  conduktor template Topic                      # Get embedded template for Topic
  conduktor template --live                     # List kinds from server
  conduktor template Topic --live               # Get template from server
  conduktor template Topic --live --cluster gw1 # Get template for specific cluster
  conduktor template Topic -o topic.yaml        # Save to file
`,
	Args: cobra.MaximumNArgs(1),
}

func initTemplate(rootContext cli.RootContext) {
	rootCmd.AddCommand(templateCmd)
	var file *string
	var edit *bool
	var apply *bool
	var live *bool
	var cluster *string

	file = templateCmd.PersistentFlags().StringP("output", "o", "", "Write example to file")
	edit = templateCmd.PersistentFlags().BoolP("edit", "e", false, "Edit the YAML file post-creation; this works only with --output. It will the EDITOR environment variable or nano if not set.")
	apply = templateCmd.PersistentFlags().BoolP("apply", "a", false, "Apply the YAML file post-editing; this works only with --edit.")
	live = templateCmd.PersistentFlags().Bool("live", false, "Fetch template from server instead of using embedded defaults")
	cluster = templateCmd.PersistentFlags().String("cluster", "", "Specify cluster for template (only used with --live)")

	templateCmd.Run = func(cmd *cobra.Command, args []string) {
		if *live {
			runTemplateLive(rootContext, args, file, edit, apply, cluster)
		} else {
			// Original behavior: show help
			_ = cmd.Help()
			os.Exit(1)
		}
	}

	// Add all kinds to the 'template' command
	for name, kind := range rootContext.Catalog.Kind {
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
				if cluster != nil && *cluster != "" && (live == nil || !*live) {
					fmt.Fprintln(os.Stderr, "Cannot use --cluster without --live")
					os.Exit(12)
				}
			},
			Run: func(cmd *cobra.Command, args []string) {
				// --live: fetch from server
				if *live {
					runTemplateLive(rootContext, []string{name}, file, edit, apply, cluster)
					return
				}

				// Original behavior: use embedded template
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
						_, err = w.WriteString(kind.GetLatestKindVersion().GetApplyExample())
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

func runTemplateLive(rootContext cli.RootContext, args []string, file *string, edit *bool, apply *bool, cluster *string) {
	apiClient := rootContext.ConsoleAPIClient()
	httpClient := apiClient.Resty()
	baseURL := apiClient.BaseURL()
	templateURL := strings.TrimSuffix(baseURL, "/api") + "/public/v1/resources/template"

	// If no kind specified, list all available kinds
	if len(args) == 0 {
		var kinds []string
		req := httpClient.R().
			SetHeader(cli.ApiVersionHeader, cli.ApiVersion).
			SetResult(&kinds)
		if cluster != nil && *cluster != "" {
			req = req.SetQueryParam("cluster", *cluster)
		}
		resp, err := req.Get(templateURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching kinds: %s\n", err)
			os.Exit(1)
		}
		if resp.IsError() {
			if strings.Contains(resp.String(), "Unsupported API version") {
				fmt.Fprintf(os.Stderr, "API version mismatch: %s\nPlease upgrade your CLI.\n", resp.String())
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error fetching kinds: %s\n", resp.String())
			os.Exit(1)
		}

		fmt.Println("Available Kinds (from server):")
		for _, k := range kinds {
			fmt.Println("  " + k)
		}
		return
	}

	// Fetch template for specific kind
	kind := args[0]
	req := httpClient.R().
		SetHeader(cli.ApiVersionHeader, cli.ApiVersion)
	if cluster != nil && *cluster != "" {
		req = req.SetQueryParam("cluster", *cluster)
	}
	resp, err := req.Get(templateURL + "/" + kind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching template: %s\n", err)
		os.Exit(1)
	}
	if resp.IsError() {
		if strings.Contains(resp.String(), "Unsupported API version") {
			fmt.Fprintf(os.Stderr, "API version mismatch: %s\nPlease upgrade your CLI.\n", resp.String())
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error fetching template for kind '%s': %s\n", kind, resp.String())
		os.Exit(1)
	}

	example := resp.String()

	// Output template (same logic as embedded)
	if file == nil || *file == "" {
		fmt.Println("---")
		fmt.Println(example)
	} else {
		_, err := os.Stat(*file)
		if err == nil {
			fmt.Fprintf(os.Stderr, "File %s already exists. You can use conduktor template %s >> %s to append to existing file\n", *file, kind, *file)
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

func editAndApply(rootContext cli.RootContext, edit *bool, file *string, apply *bool) {
	if edit != nil && *edit {
		// Run editor on the file
		err := runEditor(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Editor error: %s\n", err)
			os.Exit(7)
		}

		recursiveFolder := false
		if apply != nil && *apply {
			runApply(rootContext, []string{*file}, recursiveFolder)
		}
	}
}
