package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/context"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/parser"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check lint output against baseline",
	Long:  `Read lint output from stdin and output only issues not in the baseline.`,
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	newIssues, err := check(os.Stdin, os.Stdout, baselineFile)
	if err != nil {
		return err
	}
	if newIssues > 0 {
		os.Exit(1)
	}
	return nil
}

func check(stdin io.Reader, stdout io.Writer, baselinePath string) (int, error) {
	p := parser.NewLineParser()
	extractor := context.NewExtractor()
	store := baseline.NewStore()

	bl, err := store.Load(baselinePath)
	if err != nil {
		return 0, fmt.Errorf("loading baseline: %w", err)
	}

	matcher := baseline.NewSessionMatcher(bl, baseline.NewExactMatcher())

	newIssues := 0
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		line := scanner.Text()
		issue, err := p.Parse(line)
		if err != nil {
			fmt.Fprintln(stdout, line)
			continue
		}

		ctx, err := extractor.Extract(issue.File, issue.Line)
		if err != nil {
			ctx = &context.Context{Lines: []string{""}, Hash: ""}
		}

		sourceLine := ""
		if len(ctx.Lines) > 0 {
			sourceLine = ctx.Lines[0]
		}

		entry := baseline.Entry{
			File:       issue.File,
			Message:    issue.Message,
			SourceLine: sourceLine,
			Count:      1,
			Fingerprints: baseline.Fingerprints{
				LineHash: ctx.Hash,
			},
		}

		if !matcher.Match(entry) {
			fmt.Fprintln(stdout, line)
			newIssues++
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("reading input: %w", err)
	}

	return newIssues, nil
}
