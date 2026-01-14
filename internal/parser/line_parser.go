package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidFormat = errors.New("invalid line format")
)

// LineParser parses lint output in file:line:column:message format
type LineParser struct {
	patterns []*regexp.Regexp
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
	}
}

// Parse parses a single line of lint output
func (p *LineParser) Parse(line string) (*Issue, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, ErrInvalidFormat
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

	return nil, ErrInvalidFormat
}
