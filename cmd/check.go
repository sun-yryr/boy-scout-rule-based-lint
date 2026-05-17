package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/context"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/diff"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/parser"
)

var (
	boyScoutPolicy string
	baseRef        string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check lint output against baseline",
	Long: `Read lint output from stdin and output only issues not in the baseline.

Boy Scout Policy:
  off (default)  - baseline suppresses errors everywhere
  file           - baseline is ignored for all errors in changed files
  hunk           - baseline is ignored for errors on changed lines
  scope          - baseline is ignored for errors in changed scopes (not yet available)`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&boyScoutPolicy, "boy-scout-policy", "off",
		"Boy Scout policy: off, file, hunk, scope")
	checkCmd.Flags().StringVar(&baseRef, "base-ref", "",
		"Git base ref for Boy Scout policy (e.g. origin/main)")
	rootCmd.AddCommand(checkCmd)
}

var validPolicies = map[string]bool{
	"off":   true,
	"file":  true,
	"hunk":  true,
	"scope": true,
}

func runCheck(cmd *cobra.Command, args []string) error {
	if !validPolicies[boyScoutPolicy] {
		return fmt.Errorf("invalid --boy-scout-policy %q: valid values are off, file, hunk, scope", boyScoutPolicy)
	}

	if boyScoutPolicy == "scope" {
		return fmt.Errorf("--boy-scout-policy=scope is not yet available")
	}

	var changeSet *diff.ChangeSet
	if boyScoutPolicy != "off" {
		if baseRef == "" {
			return fmt.Errorf("--base-ref is required when --boy-scout-policy is not 'off'")
		}
		var err error
		changeSet, err = diff.GetDiff(baseRef)
		if err != nil {
			return fmt.Errorf("computing diff against %s: %w", baseRef, err)
		}
	}

	newIssues, err := check(os.Stdin, os.Stdout, baselineFile, boyScoutPolicy, changeSet)
	if err != nil {
		return err
	}
	if newIssues > 0 {
		os.Exit(1)
	}
	return nil
}

func check(stdin io.Reader, stdout io.Writer, baselinePath string, policy string, changeSet *diff.ChangeSet) (int, error) {
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

		if changeSet != nil {
			switch policy {
			case "file":
				if changeSet.HasFile(issue.File) {
					fmt.Fprintln(stdout, line)
					newIssues++
					continue
				}
			case "hunk":
				if changeSet.HasLine(issue.File, issue.Line) {
					fmt.Fprintln(stdout, line)
					newIssues++
					continue
				}
			}
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

