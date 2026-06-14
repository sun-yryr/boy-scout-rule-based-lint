package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/context"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/diff"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/parser"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/reporter"
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
		"Boy Scout policy: off, file, hunk, scope (overrides baseline config)")
	checkCmd.Flags().StringVar(&baseRef, "base-ref", "",
		"Git base ref for Boy Scout policy (e.g. origin/main; overrides baseline config)")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	store := baseline.NewStore()
	bl, err := store.Load(baselineFile)
	if err != nil {
		return fmt.Errorf("loading baseline: %w", err)
	}

	policyChanged := cmd.Flags().Changed("boy-scout-policy")
	baseRefChanged := cmd.Flags().Changed("base-ref")

	policy, err := resolveCheckPolicy(boyScoutPolicy, policyChanged, bl)
	if err != nil {
		return err
	}

	resolvedBaseRef := resolveCheckBaseRef(baseRef, baseRefChanged, bl)

	var changeSet *diff.ChangeSet
	if policy != "off" {
		if resolvedBaseRef == "" {
			return fmt.Errorf("base-ref is required when boy-scout-policy is not 'off'")
		}
		changeSet, err = diff.GetDiff(resolvedBaseRef)
		if err != nil {
			return fmt.Errorf("computing diff against %s: %w", resolvedBaseRef, err)
		}
	}

	newIssues, err := check(os.Stdin, os.Stdout, bl, policy, changeSet)
	if err != nil {
		return err
	}
	if newIssues > 0 {
		os.Exit(1)
	}
	return nil
}

type lintChecker struct {
	parser    *parser.LineParser
	extractor *context.Extractor
	matcher   *baseline.SessionMatcher
	changeSet *diff.ChangeSet
	policy    string
	reporter  *reporter.Reporter
}

func check(stdin io.Reader, stdout io.Writer, bl *baseline.Baseline, policy string, changeSet *diff.ChangeSet) (int, error) {
	c := &lintChecker{
		parser:    parser.NewLineParser(),
		extractor: context.NewExtractor(),
		matcher:   baseline.NewSessionMatcher(bl, baseline.NewExactMatcher()),
		changeSet: changeSet,
		policy:    policy,
		reporter:  reporter.NewReporter(stdout),
	}

	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		stop, err := c.handleLine(scanner.Text())
		if stop {
			return c.reporter.NewIssues(), nil
		}
		if err != nil {
			return c.reporter.NewIssues(), err
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("reading input: %w", err)
	}

	return c.reporter.NewIssues(), nil
}

func (c *lintChecker) handleLine(line string) (stop bool, err error) {
	issue, parseErr := c.parser.Parse(line)
	if errors.Is(parseErr, parser.ErrSkipLine) {
		return false, nil
	}
	if parseErr != nil {
		return c.reporter.Report(line)
	}

	if c.changeSet != nil {
		switch c.policy {
		case "file":
			if c.changeSet.HasFile(issue.File) {
				return c.reporter.Report(line)
			}
		case "hunk":
			if c.changeSet.HasLine(issue.File, issue.Line) {
				return c.reporter.Report(line)
			}
		}
	}

	ctx, err := c.extractor.Extract(issue.File, issue.Line)
	if err != nil {
		return false, fmt.Errorf("extracting context for %s:%d: %w", issue.File, issue.Line, err)
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

	if !c.matcher.Match(entry) {
		return c.reporter.Report(line)
	}

	return false, nil
}
