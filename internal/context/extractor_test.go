package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewExtractor(t *testing.T) {
	e := NewExtractor()
	if e == nil {
		t.Fatal("NewExtractor() returned nil")
	}
}

func TestExtractor_Extract(t *testing.T) {
	content := `line 1
line 2
line 3`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		lineNum   int
		wantLines []string
	}{
		{
			name:      "first line",
			lineNum:   1,
			wantLines: []string{"line 1"},
		},
		{
			name:      "middle line",
			lineNum:   2,
			wantLines: []string{"line 2"},
		},
		{
			name:      "last line",
			lineNum:   3,
			wantLines: []string{"line 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExtractor()
			ctx, err := e.Extract(tmpFile, tt.lineNum)

			if err != nil {
				t.Fatalf("Extract() unexpected error: %v", err)
			}

			if len(ctx.Lines) != len(tt.wantLines) {
				t.Errorf("Lines count = %d, want %d", len(ctx.Lines), len(tt.wantLines))
			}

			for i, line := range tt.wantLines {
				if i < len(ctx.Lines) && ctx.Lines[i] != line {
					t.Errorf("Lines[%d] = %q, want %q", i, ctx.Lines[i], line)
				}
			}

			if ctx.Hash == "" {
				t.Error("Hash should not be empty")
			}
		})
	}
}

func TestExtractor_Extract_FileNotFound(t *testing.T) {
	e := NewExtractor()
	_, err := e.Extract("/nonexistent/file.go", 1)
	if err == nil {
		t.Error("Extract() expected error for nonexistent file, got nil")
	}
}

func TestExtractor_Extract_InvalidLineNum(t *testing.T) {
	content := `line 1
line 2
line 3`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		lineNum int
	}{
		{name: "zero line number", lineNum: 0},
		{name: "negative line number", lineNum: -1},
		{name: "line number beyond file", lineNum: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExtractor()
			_, err := e.Extract(tmpFile, tt.lineNum)
			if err == nil {
				t.Errorf("Extract() lineNum=%d: expected error, got nil", tt.lineNum)
			}
		})
	}
}

func TestExtractor_SourceLinePreservesWhitespace(t *testing.T) {
	content := "  func foo() {  \n"

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	e := NewExtractor()
	ctx, err := e.Extract(tmpFile, 1)
	if err != nil {
		t.Fatalf("Extract() unexpected error: %v", err)
	}

	want := "  func foo() {  "
	if ctx.Lines[0] != want {
		t.Errorf("Lines[0] = %q, want %q (whitespace preserved)", ctx.Lines[0], want)
	}
}

func TestExtractor_HashNormalization(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	content1 := "  func foo() {  \n    return\n  }"
	content2 := "func foo() {\nreturn\n}"

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	e := NewExtractor()

	ctx1, err := e.Extract(file1, 2)
	if err != nil {
		t.Fatalf("Extract file1 failed: %v", err)
	}

	ctx2, err := e.Extract(file2, 2)
	if err != nil {
		t.Fatalf("Extract file2 failed: %v", err)
	}

	// After normalization (trimming whitespace), the hashes for the same line should match
	if ctx1.Hash != ctx2.Hash {
		t.Errorf("Hashes should match after normalization: %q != %q", ctx1.Hash, ctx2.Hash)
	}
}

func TestExtractor_DifferentContentDifferentHash(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	if err := os.WriteFile(file1, []byte("content A"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content B"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	e := NewExtractor()

	ctx1, _ := e.Extract(file1, 1)
	ctx2, _ := e.Extract(file2, 1)

	if ctx1.Hash == ctx2.Hash {
		t.Error("Different content should produce different hashes")
	}
}
