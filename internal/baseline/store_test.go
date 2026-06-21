package baseline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStore_Save_NoHTMLEscape(t *testing.T) {
	bl := New()
	bl.Entries = []Entry{{
		File:       "foo.go",
		Message:    "comparison",
		SourceLine: "if x > 0 && y < 1 {",
		Count:      1,
		Fingerprints: Fingerprints{
			LineHash: "abc123",
		},
	}}

	path := filepath.Join(t.TempDir(), "baseline.json")
	store := NewStore()
	if err := store.Save(path, bl); err != nil {
		t.Fatalf("Save() err = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() err = %v", err)
	}
	content := string(data)
	if strings.Contains(content, `\u003e`) || strings.Contains(content, `\u003c`) {
		t.Errorf("baseline JSON should not HTML-escape angle brackets: %s", content)
	}
	if !strings.Contains(content, "if x > 0 && y < 1 {") {
		t.Errorf("expected literal angle brackets in source_line: %s", content)
	}
}

func TestStore_Save_RoundTrip(t *testing.T) {
	bl := New()
	bl.Entries = []Entry{{
		File:       "main.go",
		Message:    "issue",
		SourceLine: "\tindented line",
		Count:      1,
		Fingerprints: Fingerprints{
			LineHash: "deadbeef",
		},
	}}

	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")
	store := NewStore()
	if err := store.Save(path, bl); err != nil {
		t.Fatalf("Save() err = %v", err)
	}

	loaded, err := store.Load(path)
	if err != nil {
		t.Fatalf("Load() err = %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(loaded.Entries))
	}
	if loaded.Entries[0].SourceLine != "\tindented line" {
		t.Errorf("SourceLine = %q, want preserved whitespace", loaded.Entries[0].SourceLine)
	}
}
