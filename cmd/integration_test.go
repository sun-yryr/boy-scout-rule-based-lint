package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	entries, err := os.ReadDir("testdata/scenarios")
	if err != nil {
		t.Fatalf("ReadDir(testdata/scenarios): %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			runScenario(t, filepath.Join("testdata/scenarios", entry.Name()))
		})
	}
}

func runScenario(t *testing.T, dir string) {
	t.Helper()

	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("Abs(%q): %v", dir, err)
	}

	baselineInput := readScenarioFile(t, filepath.Join(absDir, "baseline.txt"))
	checkInput := readScenarioFile(t, filepath.Join(absDir, "check.txt"))
	expected := readScenarioFile(t, filepath.Join(absDir, "expected.txt"))

	workDir := t.TempDir()
	copyScenarioSources(t, absDir, workDir, "before")
	chdirTo(t, workDir)

	baselinePath := filepath.Join(t.TempDir(), "baseline.json")
	n, err := initBaseline(strings.NewReader(baselineInput), baselinePath)
	if err != nil {
		t.Fatalf("initBaseline(): %v", err)
	}
	if baselineInput != "" && n == 0 {
		t.Fatal("initBaseline() created 0 entries for non-empty baseline input")
	}

	if hasScenarioDir(absDir, "after") {
		copyScenarioSources(t, absDir, workDir, "after")
	}

	var stdout bytes.Buffer
	newCount, err := check(strings.NewReader(checkInput), &stdout, baselinePath, "off", nil)
	if err != nil {
		t.Fatalf("check(): %v", err)
	}

	if stdout.String() != expected {
		t.Fatalf("check() stdout = %q, want %q", stdout.String(), expected)
	}

	if expected == "" {
		if newCount != 0 {
			t.Fatalf("check() new issues = %d, want 0 for all suppressed scenario", newCount)
		}
		return
	}

	if newCount <= 0 {
		t.Fatalf("check() new issues = %d, want > 0 for scenario with expected output", newCount)
	}
}

func readScenarioFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(data)
}

func hasScenarioDir(scenarioDir, phase string) bool {
	info, err := os.Stat(filepath.Join(scenarioDir, phase))
	return err == nil && info.IsDir()
}

func copyScenarioSources(t *testing.T, scenarioDir, destDir, phase string) {
	t.Helper()

	phaseDir := filepath.Join(scenarioDir, phase)
	if phase != "before" && phase != "after" {
		t.Fatalf("unsupported phase %q", phase)
	}

	if _, err := os.Stat(phaseDir); err == nil {
		copyDirFiles(t, phaseDir, destDir)
		return
	}

	if phase == "before" || !hasScenarioDir(scenarioDir, "before") {
		copyRootSources(t, scenarioDir, destDir)
	}
}

func copyRootSources(t *testing.T, scenarioDir, destDir string) {
	t.Helper()

	entries, err := os.ReadDir(scenarioDir)
	if err != nil {
		t.Fatalf("ReadDir(%q): %v", scenarioDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if isScenarioMetadata(name) {
			continue
		}
		copyFile(t, filepath.Join(scenarioDir, name), filepath.Join(destDir, name))
	}
}

func copyDirFiles(t *testing.T, srcDir, destDir string) {
	t.Helper()

	err := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		copyFile(t, path, filepath.Join(destDir, rel))
		return nil
	})
	if err != nil {
		t.Fatalf("copy from %q: %v", srcDir, err)
	}
}

func copyFile(t *testing.T, src, dest string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", filepath.Dir(dest), err)
	}

	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("Open(%q): %v", src, err)
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		t.Fatalf("Create(%q): %v", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		t.Fatalf("Copy(%q -> %q): %v", src, dest, err)
	}
}

func isScenarioMetadata(name string) bool {
	switch name {
	case "README.md", "baseline.txt", "check.txt", "expected.txt":
		return true
	default:
		return false
	}
}
