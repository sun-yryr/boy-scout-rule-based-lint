package parser

import (
	"errors"
	"testing"
)

func TestLineParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantFile    string
		wantLine    int
		wantColumn  int
		wantMessage string
		wantErr     bool
	}{
		{
			name:        "standard format with column",
			input:       "main.go:10:5: undefined: foo",
			wantFile:    "main.go",
			wantLine:    10,
			wantColumn:  5,
			wantMessage: "undefined: foo",
			wantErr:     false,
		},
		{
			name:        "standard format without column",
			input:       "main.go:42: missing return",
			wantFile:    "main.go",
			wantLine:    42,
			wantColumn:  0,
			wantMessage: "missing return",
			wantErr:     false,
		},
		{
			name:        "visual studio format with column",
			input:       "src/app.cs(15,8): error CS0001: syntax error",
			wantFile:    "src/app.cs",
			wantLine:    15,
			wantColumn:  8,
			wantMessage: "error CS0001: syntax error",
			wantErr:     false,
		},
		{
			name:        "visual studio format without column",
			input:       "Program.cs(100): warning CS0168: unused variable",
			wantFile:    "Program.cs",
			wantLine:    100,
			wantColumn:  0,
			wantMessage: "warning CS0168: unused variable",
			wantErr:     false,
		},
		{
			name:        "path with directory",
			input:       "internal/parser/parser.go:25:10: exported function should have comment",
			wantFile:    "internal/parser/parser.go",
			wantLine:    25,
			wantColumn:  10,
			wantMessage: "exported function should have comment",
			wantErr:     false,
		},
		{
			name:        "message with extra colons",
			input:       "config.go:1:1: import \"fmt\" is not used",
			wantFile:    "config.go",
			wantLine:    1,
			wantColumn:  1,
			wantMessage: "import \"fmt\" is not used",
			wantErr:     false,
		},
		{
			name:    "empty line",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   \t  ",
			wantErr: true,
		},
		{
			name:    "no colon separator",
			input:   "this is not a lint message",
			wantErr: true,
		},
		{
			name:        "extra whitespace around message",
			input:       "file.go:10:5:   extra spaces   ",
			wantFile:    "file.go",
			wantLine:    10,
			wantColumn:  5,
			wantMessage: "extra spaces",
			wantErr:     false,
			// Note: Raw will be trimmed by Parse()
		},
	}

	parser := NewLineParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue, err := parser.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if issue.File != tt.wantFile {
				t.Errorf("File = %q, want %q", issue.File, tt.wantFile)
			}
			if issue.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", issue.Line, tt.wantLine)
			}
			if issue.Column != tt.wantColumn {
				t.Errorf("Column = %d, want %d", issue.Column, tt.wantColumn)
			}
			if issue.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", issue.Message, tt.wantMessage)
			}
			// Raw stores the trimmed input line
			if issue.Raw == "" {
				t.Error("Raw should not be empty")
			}
		})
	}
}

func TestLineParser_ParseESLintStylish(t *testing.T) {
	input := []string{
		"/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
		"  1:10  error    'addOne' is defined but never used                              no-unused-vars",
		"  2:9   error    Use the isNaN function to compare with NaN                      use-isnan",
		"  3:16  error    Unexpected space before unary operator '++'                     space-unary-ops",
		"  3:16  error    The value assigned to 'i' is not used in subsequent statements  no-useless-assignment",
		"  3:20  warning  Missing semicolon                                               semi",
		"  4:12  warning  Unnecessary 'else' after 'return'                               no-else-return",
		"  5:1   warning  Expected indentation of 8 spaces but found 6                    indent",
		"  5:7   error    Function 'addOne' expected a return value                       consistent-return",
		"  5:13  warning  Missing semicolon                                               semi",
		"✖ 9 problems (5 errors, 4 warnings)",
		"  1 error and 4 warnings potentially fixable with the `--fix` option.",
	}

	wantIssues := []Issue{
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    1,
			Column:  10,
			Message: "'addOne' is defined but never used (no-unused-vars)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    2,
			Column:  9,
			Message: "Use the isNaN function to compare with NaN (use-isnan)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    3,
			Column:  16,
			Message: "Unexpected space before unary operator '++' (space-unary-ops)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    3,
			Column:  16,
			Message: "The value assigned to 'i' is not used in subsequent statements (no-useless-assignment)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    3,
			Column:  20,
			Message: "Missing semicolon (semi)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    4,
			Column:  12,
			Message: "Unnecessary 'else' after 'return' (no-else-return)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    5,
			Column:  1,
			Message: "Expected indentation of 8 spaces but found 6 (indent)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    5,
			Column:  7,
			Message: "Function 'addOne' expected a return value (consistent-return)",
		},
		{
			File:    "/var/lib/jenkins/workspace/eslint Release/eslint/fullOfProblems.js",
			Line:    5,
			Column:  13,
			Message: "Missing semicolon (semi)",
		},
	}

	p := NewLineParser()
	var gotIssues []Issue

	for _, line := range input {
		issue, err := p.Parse(line)
		if errors.Is(err, ErrSkipLine) {
			continue
		}
		if err != nil {
			t.Fatalf("Parse(%q) unexpected error: %v", line, err)
		}
		gotIssues = append(gotIssues, *issue)
	}

	if len(gotIssues) != len(wantIssues) {
		t.Fatalf("got %d issues, want %d", len(gotIssues), len(wantIssues))
	}

	for i, want := range wantIssues {
		got := gotIssues[i]
		if got.File != want.File {
			t.Errorf("issue[%d].File = %q, want %q", i, got.File, want.File)
		}
		if got.Line != want.Line {
			t.Errorf("issue[%d].Line = %d, want %d", i, got.Line, want.Line)
		}
		if got.Column != want.Column {
			t.Errorf("issue[%d].Column = %d, want %d", i, got.Column, want.Column)
		}
		if got.Message != want.Message {
			t.Errorf("issue[%d].Message = %q, want %q", i, got.Message, want.Message)
		}
	}
}

func TestLineParser_ParseESLintStylishSkipLines(t *testing.T) {
	p := NewLineParser()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "file path line",
			input: "src/app.js",
		},
		{
			name:  "summary line",
			input: "✖ 9 problems (5 errors, 4 warnings)",
		},
		{
			name:  "fixable summary line",
			input: "  1 error and 4 warnings potentially fixable with the `--fix` option.",
		},
		{
			name:  "empty line",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.input)
			if !errors.Is(err, ErrSkipLine) {
				t.Errorf("Parse() error = %v, want ErrSkipLine", err)
			}
		})
	}
}

func TestLineParser_ParseESLintStylishRelativePath(t *testing.T) {
	p := NewLineParser()

	_, err := p.Parse("src/app.js")
	if !errors.Is(err, ErrSkipLine) {
		t.Fatalf("file path line: error = %v, want ErrSkipLine", err)
	}

	issue, err := p.Parse("  1:10  error    'foo' is not defined    no-undef")
	if err != nil {
		t.Fatalf("issue line: unexpected error: %v", err)
	}

	if issue.File != "src/app.js" {
		t.Errorf("File = %q, want %q", issue.File, "src/app.js")
	}
	if issue.Message != "'foo' is not defined (no-undef)" {
		t.Errorf("Message = %q, want %q", issue.Message, "'foo' is not defined (no-undef)")
	}
}

func TestLineParser_ParseESLintUnix(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantFile    string
		wantLine    int
		wantColumn  int
		wantMessage string
	}{
		{
			name:        "classic unix format",
			input:       "src/app.js:1:10: 'addOne' is defined but never used. (no-unused-vars)",
			wantFile:    "src/app.js",
			wantLine:    1,
			wantColumn:  10,
			wantMessage: "'addOne' is defined but never used. (no-unused-vars)",
		},
		{
			name:        "eslint-formatter-unix package format",
			input:       "foo.js:5:10: Unexpected foo. [Error/foo]",
			wantFile:    "foo.js",
			wantLine:    5,
			wantColumn:  10,
			wantMessage: "Unexpected foo. [Error/foo]",
		},
	}

	p := NewLineParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}

			if issue.File != tt.wantFile {
				t.Errorf("File = %q, want %q", issue.File, tt.wantFile)
			}
			if issue.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", issue.Line, tt.wantLine)
			}
			if issue.Column != tt.wantColumn {
				t.Errorf("Column = %d, want %d", issue.Column, tt.wantColumn)
			}
			if issue.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", issue.Message, tt.wantMessage)
			}
		})
	}
}

func TestLineParser_ParseInvalidLineNotTreatedAsFilePath(t *testing.T) {
	p := NewLineParser()

	_, err := p.Parse("this is not a lint message")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("Parse() error = %v, want ErrInvalidFormat", err)
	}
}
