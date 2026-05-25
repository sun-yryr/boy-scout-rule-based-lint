package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidFormat = errors.New("invalid line format")
	ErrSkipLine      = errors.New("skip line")
)

// LineParser parses lint output in file:line:column:message format
type LineParser struct {
	patterns       []*regexp.Regexp
	eslintIssue    *regexp.Regexp
	eslintSummary  []*regexp.Regexp
	filePathSuffix *regexp.Regexp
	currentFile    string
}

// NewLineParser creates a new LineParser
func NewLineParser() *LineParser {
	return &LineParser{
		patterns: []*regexp.Regexp{
			// file:line:column: message (most common)
			regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(.+)$`),
			// file:line: message (without column)
			regexp.MustCompile(`^(.+?):(\d+):\s*(.+)$`),
			// file(line,column): message (Visual Studio style)
			regexp.MustCompile(`^(.+?)\((\d+),(\d+)\):\s*(.+)$`),
			// file(line): message (Visual Studio style without column)
			regexp.MustCompile(`^(.+?)\((\d+)\):\s*(.+)$`),
		},
		eslintIssue: regexp.MustCompile(`^(\d+):(\d+)\s+(?:error|warning)\s+(.+?\S)\s{2,}(\S+)\s*$`),
		eslintSummary: []*regexp.Regexp{
			regexp.MustCompile(`^\s*✖\s+\d+\s+problems?`),
			regexp.MustCompile(`potentially fixable`),
		},
		filePathSuffix: regexp.MustCompile(`\.\w{1,10}$`),
	}
}

// Parse parses a single line of lint output
func (p *LineParser) Parse(line string) (*Issue, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, ErrSkipLine
	}

	// Try each pattern
	for i, pattern := range p.patterns {
		matches := pattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		issue := &Issue{Raw: line}

		switch i {
		case 0: // file:line:column: message
			issue.File = matches[1]
			issue.Line, _ = strconv.Atoi(matches[2])
			issue.Column, _ = strconv.Atoi(matches[3])
			issue.Message = strings.TrimSpace(matches[4])
		case 1: // file:line: message
			issue.File = matches[1]
			issue.Line, _ = strconv.Atoi(matches[2])
			issue.Column = 0
			issue.Message = strings.TrimSpace(matches[3])
		case 2: // file(line,column): message
			issue.File = matches[1]
			issue.Line, _ = strconv.Atoi(matches[2])
			issue.Column, _ = strconv.Atoi(matches[3])
			issue.Message = strings.TrimSpace(matches[4])
		case 3: // file(line): message
			issue.File = matches[1]
			issue.Line, _ = strconv.Atoi(matches[2])
			issue.Column = 0
			issue.Message = strings.TrimSpace(matches[3])
		}

		return issue, nil
	}

	if p.currentFile != "" {
		if matches := p.eslintIssue.FindStringSubmatch(line); matches != nil {
			lineNum, _ := strconv.Atoi(matches[1])
			column, _ := strconv.Atoi(matches[2])
			message := strings.TrimSpace(matches[3])
			ruleID := matches[4]

			return &Issue{
				File:    p.currentFile,
				Line:    lineNum,
				Column:  column,
				Message: fmt.Sprintf("%s (%s)", message, ruleID),
				Raw:     line,
			}, nil
		}
	}

	for _, pattern := range p.eslintSummary {
		if pattern.MatchString(line) {
			return nil, ErrSkipLine
		}
	}

	if p.isESLintFilePathLine(line) {
		p.currentFile = line
		return nil, ErrSkipLine
	}

	return nil, ErrInvalidFormat
}

func (p *LineParser) isESLintFilePathLine(line string) bool {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return false
	}
	if strings.Contains(line, "/") || strings.Contains(line, `\`) {
		return true
	}
	return p.filePathSuffix.MatchString(line)
}
