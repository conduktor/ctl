package cmd

import (
	"fmt"
	"os"

	"text/tabwriter"

	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

func initSql(kinds schema.KindCatalog) {
	_, ok := kinds["IndexedTopic"]
	if ok {
		numLine := 1
		var sqlCmd = &cobra.Command{
			Use:  "sql",
			Args: cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				sqlResult, err := consoleApiClient().ExecuteSql(numLine, args[0])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not execute SQL: %s\n", err)
					os.Exit(1)
				}

				minwidth := 0
				tabwidth := 2
				padding := 2
				padchar := byte(' ')
				flags := uint(0)
				writer := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, flags)
				headerStr := ""
				for _, column := range sqlResult.Header {
					if headerStr != "" {
						headerStr += "\t" + column
					} else {
						headerStr += column
					}
				}
				fmt.Fprintln(writer, headerStr)
				for _, line := range sqlResult.Row {
					lineStr := ""
					for _, data := range line {
						dataStr := fmt.Sprintf("%v", data)
						if lineStr != "" {
							lineStr += "\t" + dataStr
						} else {
							lineStr += dataStr
						}
					}
					fmt.Fprintln(writer, lineStr)
				}
				writer.Flush()
			},
		}

		sqlCmd.Flags().IntVarP(&numLine, "num-line", "n", 100, "Number of line to display")
		rootCmd.AddCommand(sqlCmd)
	}
}
