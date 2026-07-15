package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
)

func TestPromptExistingBaselineFrom_DeclinesOverwrite(t *testing.T) {
	input := strings.NewReader("n\n")
	var promptOut bytes.Buffer

	overwrite, inherit, err := promptExistingBaselineFrom(input, &promptOut, ".bsr-baseline.json")
	if err != nil {
		t.Fatalf("promptExistingBaselineFrom() err = %v", err)
	}
	if overwrite {
		t.Fatal("overwrite = true, want false")
	}
	if inherit {
		t.Fatal("inherit = true, want false")
	}
}

func TestPromptExistingBaselineFrom_InheritsByDefault(t *testing.T) {
	input := strings.NewReader("y\n\n")
	var promptOut bytes.Buffer

	overwrite, inherit, err := promptExistingBaselineFrom(input, &promptOut, ".bsr-baseline.json")
	if err != nil {
		t.Fatalf("promptExistingBaselineFrom() err = %v", err)
	}
	if !overwrite {
		t.Fatal("overwrite = false, want true")
	}
	if !inherit {
		t.Fatal("inherit = false, want true")
	}
}

func TestPromptExistingBaselineFrom_DeclinesInheritance(t *testing.T) {
	input := strings.NewReader("y\nn\n")
	var promptOut bytes.Buffer

	overwrite, inherit, err := promptExistingBaselineFrom(input, &promptOut, ".bsr-baseline.json")
	if err != nil {
		t.Fatalf("promptExistingBaselineFrom() err = %v", err)
	}
	if !overwrite {
		t.Fatal("overwrite = false, want true")
	}
	if inherit {
		t.Fatal("inherit = true, want false")
	}
}

func TestPromptInitConfigFrom_DeclinesConfig(t *testing.T) {
	input := strings.NewReader("n\n")
	var promptOut bytes.Buffer

	cfg, ok, err := promptInitConfigFrom(input, &promptOut)
	if err != nil {
		t.Fatalf("promptInitConfigFrom() err = %v", err)
	}
	if ok {
		t.Fatal("promptInitConfigFrom() ok = true, want false")
	}
	if cfg != nil {
		t.Fatalf("promptInitConfigFrom() cfg = %#v, want nil", cfg)
	}
}

func TestPromptInitConfigFrom_DefaultHunkPolicy(t *testing.T) {
	input := strings.NewReader("y\n\n\n")
	var promptOut bytes.Buffer

	cfg, ok, err := promptInitConfigFrom(input, &promptOut)
	if err != nil {
		t.Fatalf("promptInitConfigFrom() err = %v", err)
	}
	if !ok {
		t.Fatal("promptInitConfigFrom() ok = false, want true")
	}
	want := &baseline.Config{
		BoyScoutPolicy: "hunk",
		BaseRef:        "origin/main",
	}
	if cfg.BoyScoutPolicy != want.BoyScoutPolicy || cfg.BaseRef != want.BaseRef {
		t.Fatalf("promptInitConfigFrom() cfg = %#v, want %#v", cfg, want)
	}
}

func TestPromptInitConfigFrom_OffPolicy(t *testing.T) {
	input := strings.NewReader("y\noff\n")
	var promptOut bytes.Buffer

	cfg, ok, err := promptInitConfigFrom(input, &promptOut)
	if err != nil {
		t.Fatalf("promptInitConfigFrom() err = %v", err)
	}
	if !ok {
		t.Fatal("promptInitConfigFrom() ok = false, want true")
	}
	if cfg.BoyScoutPolicy != "off" {
		t.Fatalf("cfg.BoyScoutPolicy = %q, want off", cfg.BoyScoutPolicy)
	}
	if cfg.BaseRef != "" {
		t.Fatalf("cfg.BaseRef = %q, want empty", cfg.BaseRef)
	}
}

func TestInitBaseline_WithConfig(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	input := "main.go:1:1: package main has no comments"

	origPrompt := initConfigPrompt
	initConfigPrompt = func(promptOut io.Writer) (*baseline.Config, bool, error) {
		return &baseline.Config{
			BoyScoutPolicy: "file",
			BaseRef:        "origin/main",
		}, true, nil
	}
	t.Cleanup(func() {
		initConfigPrompt = origPrompt
	})

	n, err := initBaseline(strings.NewReader(input), baselinePath, io.Discard)
	if err != nil {
		t.Fatalf("initBaseline() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("initBaseline() = %d entries, want 1", n)
	}

	bl := loadBaseline(t, baselinePath)
	if bl.Config == nil {
		t.Fatal("baseline config is nil")
	}
	if bl.Config.BoyScoutPolicy != "file" {
		t.Errorf("config.boy_scout_policy = %q, want file", bl.Config.BoyScoutPolicy)
	}
	if bl.Config.BaseRef != "origin/main" {
		t.Errorf("config.base_ref = %q, want origin/main", bl.Config.BaseRef)
	}
}

func TestInitBaseline_DeclinesOverwrite(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	original := []byte(`{"version":2,"entries":[]}`)
	if err := os.WriteFile(baselinePath, original, 0o644); err != nil {
		t.Fatalf("WriteFile(baseline): %v", err)
	}

	origExistingPrompt := initExistingBaselinePrompt
	initExistingBaselinePrompt = func(string, io.Writer) (bool, bool, error) {
		return false, false, nil
	}
	t.Cleanup(func() {
		initExistingBaselinePrompt = origExistingPrompt
	})

	_, err := initBaseline(
		strings.NewReader("main.go:1:1: package main has no comments"),
		baselinePath,
		io.Discard,
	)
	if !errors.Is(err, errInitCanceled) {
		t.Fatalf("initBaseline() err = %v, want errInitCanceled", err)
	}

	got, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("ReadFile(baseline): %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("baseline was modified: got %q, want %q", got, original)
	}
}

func TestInitBaseline_OverwritesInvalidBaselineWithoutInheritance(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	if err := os.WriteFile(baselinePath, []byte("invalid"), 0o644); err != nil {
		t.Fatalf("WriteFile(baseline): %v", err)
	}

	origExistingPrompt := initExistingBaselinePrompt
	origConfigPrompt := initConfigPrompt
	initExistingBaselinePrompt = func(string, io.Writer) (bool, bool, error) {
		return true, false, nil
	}
	initConfigPrompt = func(io.Writer) (*baseline.Config, bool, error) {
		return nil, false, nil
	}
	t.Cleanup(func() {
		initExistingBaselinePrompt = origExistingPrompt
		initConfigPrompt = origConfigPrompt
	})

	_, err := initBaseline(
		strings.NewReader("main.go:1:1: package main has no comments"),
		baselinePath,
		io.Discard,
	)
	if err != nil {
		t.Fatalf("initBaseline() err = %v", err)
	}

	bl := loadBaseline(t, baselinePath)
	if bl.Len() != 1 {
		t.Fatalf("baseline entries = %d, want 1", bl.Len())
	}
}

func TestInitBaseline_InheritsExistingConfig(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	store := baseline.NewStore()
	existing := baseline.New()
	existing.Config = &baseline.Config{
		BoyScoutPolicy: "file",
		BaseRef:        "origin/main",
	}
	if err := store.Save(baselinePath, existing); err != nil {
		t.Fatalf("Save(existing baseline): %v", err)
	}

	origExistingPrompt := initExistingBaselinePrompt
	origConfigPrompt := initConfigPrompt
	initExistingBaselinePrompt = func(string, io.Writer) (bool, bool, error) {
		return true, true, nil
	}
	initConfigPrompt = func(io.Writer) (*baseline.Config, bool, error) {
		t.Fatal("initConfigPrompt() called while inheriting settings")
		return nil, false, nil
	}
	t.Cleanup(func() {
		initExistingBaselinePrompt = origExistingPrompt
		initConfigPrompt = origConfigPrompt
	})

	_, err := initBaseline(
		strings.NewReader("main.go:1:1: package main has no comments"),
		baselinePath,
		io.Discard,
	)
	if err != nil {
		t.Fatalf("initBaseline() err = %v", err)
	}

	bl := loadBaseline(t, baselinePath)
	if bl.Config == nil {
		t.Fatal("baseline config is nil")
	}
	if bl.Config.BoyScoutPolicy != "file" || bl.Config.BaseRef != "origin/main" {
		t.Fatalf("baseline config = %#v, want inherited config", bl.Config)
	}
}

func TestInitBaseline_ReconfiguresWhenInheritanceDeclined(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(workDir, "baseline.json")
	store := baseline.NewStore()
	existing := baseline.New()
	existing.Config = &baseline.Config{
		BoyScoutPolicy: "file",
		BaseRef:        "origin/main",
	}
	if err := store.Save(baselinePath, existing); err != nil {
		t.Fatalf("Save(existing baseline): %v", err)
	}

	origExistingPrompt := initExistingBaselinePrompt
	origConfigPrompt := initConfigPrompt
	initExistingBaselinePrompt = func(string, io.Writer) (bool, bool, error) {
		return true, false, nil
	}
	initConfigPrompt = func(io.Writer) (*baseline.Config, bool, error) {
		return &baseline.Config{BoyScoutPolicy: "off"}, true, nil
	}
	t.Cleanup(func() {
		initExistingBaselinePrompt = origExistingPrompt
		initConfigPrompt = origConfigPrompt
	})

	_, err := initBaseline(
		strings.NewReader("main.go:1:1: package main has no comments"),
		baselinePath,
		io.Discard,
	)
	if err != nil {
		t.Fatalf("initBaseline() err = %v", err)
	}

	bl := loadBaseline(t, baselinePath)
	if bl.Config == nil || bl.Config.BoyScoutPolicy != "off" {
		t.Fatalf("baseline config = %#v, want reconfigured off policy", bl.Config)
	}
}
