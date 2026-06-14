package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
)

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
