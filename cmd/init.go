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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize baseline from lint output",
	Long:  `Read lint output from stdin and create a new baseline file.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	n, err := initBaseline(os.Stdin, baselineFile)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Baseline created with %d entries\n", n)
	return nil
}

func initBaseline(stdin io.Reader, baselinePath string) (int, error) {
	p := parser.NewLineParser()
	extractor := context.NewExtractor()
	store := baseline.NewStore()

	bl := baseline.New()

	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		issue, err := p.Parse(line)
		if err != nil {
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

		bl.Add(entry)
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("reading input: %w", err)
	}

	if err := store.Save(baselinePath, bl); err != nil {
		return 0, fmt.Errorf("saving baseline: %w", err)
	}

	return bl.Len(), nil
}
