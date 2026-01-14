package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	baselineFile string
	contextLines int
)

var rootCmd = &cobra.Command{
	Use:   "bsr",
	Short: "Boy Scout Rule - lint baseline filter",
	Long: `bsr is a tool that filters lint output based on a baseline.
It allows you to ignore existing errors and only report new ones,
following the Boy Scout Rule: "Leave it better than you found it."`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&baselineFile, "baseline", "b", ".bsr-baseline.json", "baseline file path")
	rootCmd.PersistentFlags().IntVarP(&contextLines, "context", "c", 2, "number of context lines for matching")
}
