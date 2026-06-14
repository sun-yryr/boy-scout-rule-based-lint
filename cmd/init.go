package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/context"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/parser"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize baseline from lint output",
	Long: `Read lint output from stdin and create a new baseline file.

When run from an interactive terminal, bsr prompts for optional Boy Scout
Policy settings to store in the baseline config.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	n, err := initBaseline(os.Stdin, baselineFile, os.Stderr)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Baseline created with %d entries\n", n)
	return nil
}

func initBaseline(stdin io.Reader, baselinePath string, promptOut io.Writer) (int, error) {
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
			return 0, fmt.Errorf("extracting context for %s:%d: %w", issue.File, issue.Line, err)
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

	if cfg, ok, err := initConfigPrompt(promptOut); err != nil {
		return 0, err
	} else if ok {
		bl.Config = cfg
	}

	if err := store.Save(baselinePath, bl); err != nil {
		return 0, fmt.Errorf("saving baseline: %w", err)
	}

	return bl.Len(), nil
}

var initConfigPrompt = defaultInitConfigPrompt

func defaultInitConfigPrompt(promptOut io.Writer) (*baseline.Config, bool, error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return nil, false, nil
	}
	defer tty.Close()

	return promptInitConfigFrom(tty, promptOut)
}

func promptInitConfigFrom(reader io.Reader, promptOut io.Writer) (*baseline.Config, bool, error) {
	bufReader := bufio.NewReader(reader)

	configure, err := promptYesNo(bufReader, promptOut, "Configure Boy Scout Policy?", false)
	if err != nil {
		return nil, false, err
	}
	if !configure {
		return nil, false, nil
	}

	policy, err := promptPolicy(bufReader, promptOut)
	if err != nil {
		return nil, false, err
	}
	if policy == "off" {
		return &baseline.Config{BoyScoutPolicy: "off"}, true, nil
	}

	baseRef, err := promptBaseRef(bufReader, promptOut)
	if err != nil {
		return nil, false, err
	}

	return &baseline.Config{
		BoyScoutPolicy: policy,
		BaseRef:        baseRef,
	}, true, nil
}

func promptYesNo(reader *bufio.Reader, out io.Writer, question string, defaultYes bool) (bool, error) {
	defaultLabel := "y/N"
	if defaultYes {
		defaultLabel = "Y/n"
	}

	for {
		fmt.Fprintf(out, "%s [%s]: ", question, defaultLabel)
		line, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("reading input: %w", err)
		}

		answer := strings.TrimSpace(strings.ToLower(line))
		if answer == "" {
			return defaultYes, nil
		}
		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(out, "Please answer y or n.")
		}
	}
}

func promptPolicy(reader *bufio.Reader, out io.Writer) (string, error) {
	for {
		fmt.Fprint(out, "Policy [file/hunk/off] (default: hunk): ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}

		answer := strings.TrimSpace(strings.ToLower(line))
		if answer == "" {
			return "hunk", nil
		}
		if answer == "scope" {
			fmt.Fprintln(out, "scope is not yet available; choose file, hunk, or off.")
			continue
		}
		if err := validatePolicy(answer); err != nil {
			fmt.Fprintf(out, "%v\n", err)
			continue
		}
		return answer, nil
	}
}

func promptBaseRef(reader *bufio.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "Base ref (default: origin/main): ")
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	answer := strings.TrimSpace(line)
	if answer == "" {
		return "origin/main", nil
	}
	return answer, nil
}
