package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/diff"
)

func createBaseline(t *testing.T, workDir, input string) string {
	t.Helper()

	chdirTo(t, workDir)
	baselinePath := filepath.Join(workDir, "baseline.json")
	if _, err := initBaseline(strings.NewReader(input), baselinePath); err != nil {
		t.Fatalf("initBaseline(): %v", err)
	}
	return baselinePath
}

func TestCheck_AllSuppressed(t *testing.T) {
	workDir := t.TempDir()
	writeTestFile(t, workDir, "main.go", "package main\n\nfunc main() {}\n")

	input := "main.go:2:1: empty line"
	baselinePath := createBaseline(t, workDir, input)

	var out bytes.Buffer
	n, err := check(strings.NewReader(input), &out, baselinePath, "off", nil)
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
	baselinePath := createBaseline(t, workDir, baselineInput)

	checkInput := strings.Join([]string{
		baselineInput,
		"main.go:3:1: func main is unused",
	}, "\n")

	var out bytes.Buffer
	n, err := check(strings.NewReader(checkInput), &out, baselinePath, "off", nil)
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
	baselinePath := createBaseline(t, workDir, "main.go:1:1: package main has no comments")

	var out bytes.Buffer
	unparsable := "this is not a lint message"
	n, err := check(strings.NewReader(unparsable), &out, baselinePath, "off", nil)
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
	baselinePath := createBaseline(t, workDir, "main.go:1:1: package main has no comments")

	input := strings.Join([]string{
		"✖ 9 problems (5 errors, 4 warnings)",
		"main.go:1:1: package main has no comments",
	}, "\n")

	var out bytes.Buffer
	n, err := check(strings.NewReader(input), &out, baselinePath, "off", nil)
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
	baselinePath := createBaseline(t, workDir, input)

	writeTestFile(t, workDir, "main.go", "package main\nchanged content\nfunc main() {}\n")

	var out bytes.Buffer
	n, err := check(strings.NewReader(input), &out, baselinePath, "off", nil)
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
	baselinePath := createBaseline(t, workDir, baselineInput)

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
	n, err := check(strings.NewReader(checkInput), &out, baselinePath, "file", changeSet)
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
	baselinePath := createBaseline(t, workDir, baselineInput)

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
	n, err := check(strings.NewReader(baselineInput), &out, baselinePath, "hunk", changeSet)
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
	baselinePath := createBaseline(t, workDir, baselineInput)

	changeSet := mustParseDiff(t, `diff --git a/changed.go b/changed.go
--- a/changed.go
+++ b/changed.go
@@ -1 +1,2 @@
 package main
+// added
`)

	var out bytes.Buffer
	n, err := check(strings.NewReader(baselineInput), &out, baselinePath, "file", changeSet)
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
