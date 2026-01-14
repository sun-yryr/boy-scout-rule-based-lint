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

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter lint output against baseline",
	Long:  `Read lint output from stdin and output only issues not in the baseline.`,
	RunE:  runFilter,
}

func init() {
	rootCmd.AddCommand(filterCmd)
}

func runFilter(cmd *cobra.Command, args []string) error {
	p := parser.NewLineParser()
	extractor := context.NewExtractor(contextLines)
	store := baseline.NewStore()
	matcher := baseline.NewMatcher()

	bl, err := store.Load(baselineFile)
	if err != nil {
		return fmt.Errorf("loading baseline: %w", err)
	}

	newIssues := 0
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		issue, err := p.Parse(line)
		if err != nil {
			fmt.Println(line)
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
		}

		if !matcher.Match(bl, entry) {
			fmt.Println(line)
			newIssues++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	if newIssues > 0 {
		os.Exit(1)
	}

	return nil
}
