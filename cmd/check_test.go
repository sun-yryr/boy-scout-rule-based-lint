package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/diff"
)

func createBaseline(t *testing.T, workDir, input string) *baseline.Baseline {
	t.Helper()

	chdirTo(t, workDir)
	baselinePath := filepath.Join(workDir, "baseline.json")
	if _, err := initBaseline(strings.NewReader(input), baselinePath, io.Discard); err != nil {
		t.Fatalf("initBaseline(): %v", err)
	}
	return loadBaseline(t, baselinePath)
}

func TestCheck_AllSuppressed(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n\nfunc main() {}\n")

	input := "main.go:2:1: empty line"
	bl := createBaseline(t, workDir, input)

	var out bytes.Buffer
	n, err := check(strings.NewReader(input), &out, bl, "off", nil)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 0 {
		t.Fatalf("check() new issues = %d, want 0", n)
	}
	if out.Len() != 0 {
		t.Fatalf("check() stdout = %q, want empty", out.String())
	}
}

func TestCheck_NewIssue(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n\nfunc main() {}\n")

	baselineInput := "main.go:2:1: empty line"
	bl := createBaseline(t, workDir, baselineInput)

	checkInput := strings.Join([]string{
		baselineInput,
		"main.go:3:1: func main is unused",
	}, "\n")

	var out bytes.Buffer
	n, err := check(strings.NewReader(checkInput), &out, bl, "off", nil)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("check() new issues = %d, want 1", n)
	}

	want := "main.go:3:1: func main is unused\n"
	if out.String() != want {
		t.Fatalf("check() stdout = %q, want %q", out.String(), want)
	}
}

func TestCheck_UnparsablePassthrough(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	bl := createBaseline(t, workDir, "main.go:1:1: package main has no comments")

	var out bytes.Buffer
	unparsable := "this is not a lint message"
	n, err := check(strings.NewReader(unparsable), &out, bl, "off", nil)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("check() new issues = %d, want 1", n)
	}
	if out.String() != unparsable+"\n" {
		t.Fatalf("check() stdout = %q, want %q", out.String(), unparsable+"\n")
	}
}

func TestCheck_SkipLinesNotOutput(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n")
	bl := createBaseline(t, workDir, "main.go:1:1: package main has no comments")

	input := strings.Join([]string{
		"✖ 9 problems (5 errors, 4 warnings)",
		"main.go:1:1: package main has no comments",
	}, "\n")

	var out bytes.Buffer
	n, err := check(strings.NewReader(input), &out, bl, "off", nil)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 0 {
		t.Fatalf("check() new issues = %d, want 0", n)
	}
	if out.Len() != 0 {
		t.Fatalf("check() stdout = %q, want empty", out.String())
	}
}

func TestCheck_CodeChangedHashMismatch(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\nplaceholder\nfunc main() {}\n")

	input := "main.go:2:1: placeholder issue"
	bl := createBaseline(t, workDir, input)

	writeTestFile(t, workDir, "main.go", "package main\nchanged content\nfunc main() {}\n")

	var out bytes.Buffer
	n, err := check(strings.NewReader(input), &out, bl, "off", nil)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("check() new issues = %d, want 1", n)
	}
	if out.String() != input+"\n" {
		t.Fatalf("check() stdout = %q, want %q", out.String(), input+"\n")
	}
}

func TestCheck_PolicyFile_ReportsBaselineInChangedFile(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n\nfunc main() {}\n")
	writeTestFile(t, workDir, "other.go", "package main\n")

	baselineInput := strings.Join([]string{
		"main.go:2:1: empty line",
		"other.go:1:1: package main has no comments",
	}, "\n")
	bl := createBaseline(t, workDir, baselineInput)

	changeSet := mustParseDiff(t, `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
 package main
-
+// changed
 func main() {}
`)

	checkInput := baselineInput
	var out bytes.Buffer
	n, err := check(strings.NewReader(checkInput), &out, bl, "file", changeSet)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("check() new issues = %d, want 1", n)
	}
	want := "main.go:2:1: empty line\n"
	if out.String() != want {
		t.Fatalf("check() stdout = %q, want %q", out.String(), want)
	}
}

func TestCheck_PolicyHunk_ReportsOnlyChangedLine(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "line one\nline two\nline three\n")

	baselineInput := strings.Join([]string{
		"main.go:1:1: issue on line one",
		"main.go:2:1: issue on line two",
		"main.go:3:1: issue on line three",
	}, "\n")
	bl := createBaseline(t, workDir, baselineInput)

	changeSet := mustParseDiff(t, `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
 line one
-line two
+changed line two
 line three
`)

	var out bytes.Buffer
	n, err := check(strings.NewReader(baselineInput), &out, bl, "hunk", changeSet)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("check() new issues = %d, want 1", n)
	}
	want := "main.go:2:1: issue on line two\n"
	if out.String() != want {
		t.Fatalf("check() stdout = %q, want %q", out.String(), want)
	}
}

func TestCheck_PolicyFile_SuppressesUnchangedFile(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "changed.go", "package main\n")
	writeTestFile(t, workDir, "unchanged.go", "package main\n")

	baselineInput := strings.Join([]string{
		"changed.go:1:1: issue in changed file",
		"unchanged.go:1:1: issue in unchanged file",
	}, "\n")
	bl := createBaseline(t, workDir, baselineInput)

	changeSet := mustParseDiff(t, `diff --git a/changed.go b/changed.go
--- a/changed.go
+++ b/changed.go
@@ -1 +1,2 @@
 package main
+// added
`)

	var out bytes.Buffer
	n, err := check(strings.NewReader(baselineInput), &out, bl, "file", changeSet)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("check() new issues = %d, want 1", n)
	}
	want := "changed.go:1:1: issue in changed file\n"
	if out.String() != want {
		t.Fatalf("check() stdout = %q, want %q", out.String(), want)
	}
}

func mustParseDiff(t *testing.T, diffText string) *diff.ChangeSet {
	t.Helper()

	cs, err := diff.ParseDiff(strings.NewReader(diffText))
	if err != nil {
		t.Fatalf("ParseDiff(): %v", err)
	}
	return cs
}

func TestCheck_UsesBaselineConfigPolicy(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n\nfunc main() {}\n")
	writeTestFile(t, workDir, "other.go", "package main\n")

	baselineInput := strings.Join([]string{
		"main.go:2:1: empty line",
		"other.go:1:1: package main has no comments",
	}, "\n")
	bl := createBaseline(t, workDir, baselineInput)
	bl.Config = &baseline.Config{
		BoyScoutPolicy: "file",
		BaseRef:        "origin/main",
	}

	changeSet := mustParseDiff(t, `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
 package main
-
+// changed
 func main() {}
`)

	policy, err := resolveCheckPolicy("off", false, bl)
	if err != nil {
		t.Fatalf("resolveCheckPolicy() err = %v", err)
	}
	if policy != "file" {
		t.Fatalf("resolveCheckPolicy() = %q, want file", policy)
	}

	var out bytes.Buffer
	n, err := check(strings.NewReader(baselineInput), &out, bl, policy, changeSet)
	if err != nil {
		t.Fatalf("check() err = %v", err)
	}
	if n != 1 {
		t.Fatalf("check() new issues = %d, want 1", n)
	}
	want := "main.go:2:1: empty line\n"
	if out.String() != want {
		t.Fatalf("check() stdout = %q, want %q", out.String(), want)
	}
}

func TestResolveCheckPolicy_CLIOverridesBaseline(t *testing.T) {
	bl := &baseline.Baseline{
		Config: &baseline.Config{BoyScoutPolicy: "hunk"},
	}

	policy, err := resolveCheckPolicy("file", true, bl)
	if err != nil {
		t.Fatalf("resolveCheckPolicy() err = %v", err)
	}
	if policy != "file" {
		t.Fatalf("resolveCheckPolicy() = %q, want file", policy)
	}
}

func TestResolveCheckBaseRef_CLIOverridesBaseline(t *testing.T) {
	bl := &baseline.Baseline{
		Config: &baseline.Config{BaseRef: "origin/main"},
	}

	got := resolveCheckBaseRef("origin/develop", true, bl)
	if got != "origin/develop" {
		t.Fatalf("resolveCheckBaseRef() = %q, want origin/develop", got)
	}
}

func TestResolveCheckBaseRef_UsesBaselineConfig(t *testing.T) {
	bl := &baseline.Baseline{
		Config: &baseline.Config{BaseRef: "origin/main"},
	}

	got := resolveCheckBaseRef("", false, bl)
	if got != "origin/main" {
		t.Fatalf("resolveCheckBaseRef() = %q, want origin/main", got)
	}
}
