package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/context"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/parser"
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
	p := parser.NewLineParser()
	extractor := context.NewExtractor(contextLines)
	store := baseline.NewStore()

	bl := baseline.New()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		issue, err := p.Parse(line)
		if err != nil {
			continue
		}

		ctx, err := extractor.Extract(issue.File, issue.Line)
		if err != nil {
			ctx = &context.Context{Lines: []string{}, Hash: ""}
		}

		entry := baseline.Entry{
			File:         issue.File,
			Message:      issue.Message,
			ContextHash:  ctx.Hash,
			ContextLines: ctx.Lines,
			Count:        1,
		}

		bl.Add(entry)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	if err := store.Save(baselineFile, bl); err != nil {
		return fmt.Errorf("saving baseline: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Baseline updated with %d entries\n", bl.Len())
	return nil
}
