package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewExtractor(t *testing.T) {
	e := NewExtractor(3)
	if e == nil {
		t.Fatal("NewExtractor() returned nil")
	}
	if e.contextLines != 3 {
		t.Errorf("contextLines = %d, want 3", e.contextLines)
	}
}

func TestExtractor_Extract(t *testing.T) {
	// Create a temporary test file
	content := `line 1
line 2
line 3
line 4
line 5
line 6
line 7
line 8
line 9
line 10`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		contextLines int
		lineNum      int
		wantLines    []string
		wantErr      bool
	}{
		{
			name:         "middle of file with context 2",
			contextLines: 2,
			lineNum:      5,
			wantLines:    []string{"line 3", "line 4", "line 5", "line 6", "line 7"},
		},
		{
			name:         "start of file",
			contextLines: 2,
			lineNum:      1,
			wantLines:    []string{"line 1", "line 2", "line 3"},
		},
		{
			name:         "end of file",
			contextLines: 2,
			lineNum:      10,
			wantLines:    []string{"line 8", "line 9", "line 10"},
		},
		{
			name:         "context 0",
			contextLines: 0,
			lineNum:      5,
			wantLines:    []string{"line 5"},
		},
		{
			name:         "large context",
			contextLines: 20,
			lineNum:      5,
			wantLines:    []string{"line 1", "line 2", "line 3", "line 4", "line 5", "line 6", "line 7", "line 8", "line 9", "line 10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExtractor(tt.contextLines)
			ctx, err := e.Extract(tmpFile, tt.lineNum)

			if tt.wantErr {
				if err == nil {
					t.Error("Extract() expected error, got nil")
				}
				return
			}

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
	e := NewExtractor(2)
	_, err := e.Extract("/nonexistent/file.go", 1)
	if err == nil {
		t.Error("Extract() expected error for nonexistent file, got nil")
	}
}

func TestExtractor_HashNormalization(t *testing.T) {
	// Create two files with same content but different whitespace
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	// Same content with different leading/trailing whitespace
	content1 := "  func foo() {  \n    return\n  }"
	content2 := "func foo() {\nreturn\n}"

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	e := NewExtractor(0)

	ctx1, err := e.Extract(file1, 2)
	if err != nil {
		t.Fatalf("Extract file1 failed: %v", err)
	}

	ctx2, err := e.Extract(file2, 2)
	if err != nil {
		t.Fatalf("Extract file2 failed: %v", err)
	}

	// After normalization (trimming whitespace), the hashes should be the same
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

	e := NewExtractor(0)

	ctx1, _ := e.Extract(file1, 1)
	ctx2, _ := e.Extract(file2, 1)

	if ctx1.Hash == ctx2.Hash {
		t.Error("Different content should produce different hashes")
	}
}
