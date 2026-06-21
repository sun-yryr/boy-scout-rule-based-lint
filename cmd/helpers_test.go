package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func chdirTo(t *testing.T, dir string) {
	t.Helper()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q): %v", dir, err)
	}
}

func loadBaseline(t *testing.T, path string) *baseline.Baseline {
	t.Helper()

	store := baseline.NewStore()
	bl, err := store.Load(path)
	if err != nil {
		t.Fatalf("Load(%q): %v", path, err)
	}
	return bl
}
