package diff

import (
	"strings"
	"testing"
)

func TestParseDiff_SingleFileAdditions(t *testing.T) {
	diff := `diff --git a/internal/foo.go b/internal/foo.go
index abc..def 100644
--- a/internal/foo.go
+++ b/internal/foo.go
@@ -10,5 +10,7 @@ func Foo() {
 unchanged
+added line 1
 unchanged
+added line 2
 unchanged
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Files["internal/foo.go"] {
		t.Error("expected file internal/foo.go to be in changed files")
	}
	if len(cs.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(cs.Files))
	}
	lines := cs.ChangedLines["internal/foo.go"]
	if !lines[11] {
		t.Error("expected line 11 to be changed")
	}
	if !lines[13] {
		t.Error("expected line 13 to be changed")
	}
	if lines[10] {
		t.Error("line 10 should not be changed (context line)")
	}
	if len(lines) != 2 {
		t.Errorf("expected 2 changed lines, got %d", len(lines))
	}
}

func TestParseDiff_MultipleFiles(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1,1 +1,2 @@
+new line in foo
diff --git a/bar.go b/bar.go
--- a/bar.go
+++ b/bar.go
@@ -5,3 +5,4 @@
+new line in bar
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Files["foo.go"] {
		t.Error("expected foo.go")
	}
	if !cs.Files["bar.go"] {
		t.Error("expected bar.go")
	}
	if len(cs.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(cs.Files))
	}
	if !cs.ChangedLines["foo.go"][1] {
		t.Error("expected line 1 in foo.go")
	}
	if !cs.ChangedLines["bar.go"][5] {
		t.Error("expected line 5 in bar.go")
	}
}

func TestParseDiff_NewFile(t *testing.T) {
	diff := `diff --git a/newfile.go b/newfile.go
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+line 1
+line 2
+line 3
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Files["newfile.go"] {
		t.Error("expected newfile.go")
	}
	lines := cs.ChangedLines["newfile.go"]
	if len(lines) != 3 {
		t.Errorf("expected 3 changed lines, got %d", len(lines))
	}
	if !lines[1] || !lines[2] || !lines[3] {
		t.Error("expected lines 1,2,3 to be changed")
	}
}

func TestParseDiff_DeletedFile(t *testing.T) {
	diff := `diff --git a/oldfile.go b/oldfile.go
deleted file mode 100644
index abc..0000000
--- a/oldfile.go
+++ /dev/null
@@ -1,3 +0,0 @@
-line 1
-line 2
-line 3
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if cs.Files["oldfile.go"] {
		t.Error("deleted file should not be in changed files")
	}
}

func TestParseDiff_MixedAdditionsAndDeletions(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -10,5 +10,5 @@ func Bar() {
 keep
-deleted
+added
 keep
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	lines := cs.ChangedLines["foo.go"]
	if lines[10] {
		t.Error("line 10 (unchanged context) should not be changed")
	}
	if !lines[11] {
		t.Error("expected line 11 to be the added line")
	}
	if lines[12] {
		t.Error("line 12 (unchanged context) should not be changed")
	}
}

func TestParseDiff_MultipleHunks(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -5,3 +5,4 @@ func First() {
+added in first hunk
@@ -20,2 +21,3 @@ func Second() {
+added in second hunk
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	lines := cs.ChangedLines["foo.go"]
	if !lines[5] {
		t.Error("expected line 5 to be changed (first hunk)")
	}
	if !lines[21] {
		t.Error("expected line 21 to be changed (second hunk)")
	}
	if len(lines) != 2 {
		t.Errorf("expected 2 changed lines, got %d", len(lines))
	}
}

func TestParseDiff_EmptyDiff(t *testing.T) {
	cs, err := ParseDiff(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(cs.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(cs.Files))
	}
}

func TestParseDiff_HunkWithCountOmitted(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1 +1,2 @@
+added
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.ChangedLines["foo.go"][1] {
		t.Error("expected line 1 to be changed")
	}
}

func TestParseDiff_BinaryFile(t *testing.T) {
	diff := `diff --git a/image.png b/image.png
index abc..def 100644
Binary files a/image.png and b/image.png differ
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if cs.Files["image.png"] {
		t.Error("binary files should not be tracked (no hunks)")
	}
}

func TestParseDiff_NoNewlineMarker(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1,3 +1,3 @@
 line1
 line2
-line3
+line3 modified
\ No newline at end of file
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	lines := cs.ChangedLines["foo.go"]
	if lines[1] {
		t.Error("line 1 should not be changed")
	}
	if !lines[3] {
		t.Error("line 3 should be the modified line")
	}
}

func TestParseDiff_PathWithTabTimestamp(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go	2024-01-01 00:00:00.000000000 +0000
+++ b/foo.go	2024-01-01 00:00:00.000000000 +0000
@@ -1 +1,2 @@
+added
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Files["foo.go"] {
		t.Error("expected foo.go (tab+timestamp should be stripped)")
	}
}

func TestParseDiff_RenamedFile(t *testing.T) {
	diff := `diff --git a/old.go b/new.go
similarity index 100%
rename from old.go
rename to new.go
--- a/old.go
+++ b/new.go
@@ -1 +1 @@
-old
+new
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if cs.Files["old.go"] {
		t.Error("old file name should not be tracked")
	}
	if !cs.Files["new.go"] {
		t.Error("new file name should be tracked")
	}
}

func TestParseDiff_OnlyDeletions(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -5,3 +5,2 @@ func Foo() {
 context
-deleted
 context
`
	cs, err := ParseDiff(strings.NewReader(diff))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Files["foo.go"] {
		t.Error("file should be in changed files (even with only deletions)")
	}
	if len(cs.ChangedLines["foo.go"]) != 0 {
		t.Errorf("expected 0 changed lines (only deletions), got %d", len(cs.ChangedLines["foo.go"]))
	}
}
