package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
)

func TestInitBaseline_ValidLines(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n\nfunc main() {}\n")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	input := strings.Join([]string{
		"main.go:2:1: empty line",
		"main.go:3:1: func main is unused",
	}, "\n")

	n, err := initBaseline(strings.NewReader(input), baselinePath)
	if err != nil {
		t.Fatalf("initBaseline() err = %v", err)
	}
	if n != 2 {
		t.Fatalf("initBaseline() = %d entries, want 2", n)
	}

	data, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("ReadFile(baseline): %v", err)
	}

	var bl baseline.Baseline
	if err := json.Unmarshal(data, &bl); err != nil {
		t.Fatalf("Unmarshal baseline: %v", err)
	}
	if bl.Version != 2 {
		t.Errorf("baseline version = %d, want 2", bl.Version)
	}
	if len(bl.Entries) != 2 {
		t.Fatalf("baseline entries = %d, want 2", len(bl.Entries))
	}
	for _, entry := range bl.Entries {
		if entry.File != "main.go" {
			t.Errorf("entry file = %q, want main.go", entry.File)
		}
		if entry.Fingerprints.LineHash == "" {
			t.Error("entry line hash should not be empty")
		}
	}
}

func TestInitBaseline_SkipsInvalidAndSummaryLines(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	input := strings.Join([]string{
		"this is not a lint message",
		"✖ 9 problems (5 errors, 4 warnings)",
		"main.go:1:1: package main has no comments",
	}, "\n")

	n, err := initBaseline(strings.NewReader(input), baselinePath)
	if err != nil {
		t.Fatalf("initBaseline() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("initBaseline() = %d entries, want 1", n)
	}
}

func TestInitBaseline_EmptyInput(t *testing.T) {
	workDir := t.TempDir()
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	n, err := initBaseline(strings.NewReader(""), baselinePath)
	if err != nil {
		t.Fatalf("initBaseline() err = %v", err)
	}
	if n != 0 {
		t.Fatalf("initBaseline() = %d entries, want 0", n)
	}

	data, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("ReadFile(baseline): %v", err)
	}

	var bl baseline.Baseline
	if err := json.Unmarshal(data, &bl); err != nil {
		t.Fatalf("Unmarshal baseline: %v", err)
	}
	if bl.Version != 2 {
		t.Errorf("baseline version = %d, want 2", bl.Version)
	}
	if len(bl.Entries) != 0 {
		t.Errorf("baseline entries = %d, want 0", len(bl.Entries))
	}
}

func TestInitBaseline_MissingSourceFile(t *testing.T) {
	workDir := t.TempDir()
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	_, err := initBaseline(strings.NewReader("missing.go:1:1: undefined symbol"), baselinePath)
	if err == nil {
		t.Fatal("initBaseline() err = nil, want error for missing source file")
	}
}
