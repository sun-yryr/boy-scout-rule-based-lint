package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update baseline from lint output",
	Long:  `Read lint output from stdin and update the baseline file.`,
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	n, err := initBaseline(os.Stdin, baselineFile)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Baseline updated with %d entries\n", n)
	return nil
}
