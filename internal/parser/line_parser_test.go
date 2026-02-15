package parser

import (
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
