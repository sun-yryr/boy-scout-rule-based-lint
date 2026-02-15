package baseline

import (
	"testing"
)

func TestNew(t *testing.T) {
	bl := New()

	if bl.Version != 1 {
		t.Errorf("Version = %d, want 1", bl.Version)
	}
	if bl.Len() != 0 {
		t.Errorf("Len() = %d, want 0", bl.Len())
	}
}

func TestBaseline_Add(t *testing.T) {
	tests := []struct {
		name         string
		entries      []Entry
		wantLen      int
		wantCountFor map[string]int // file+message -> expected count
	}{
		{
			name: "single entry",
			entries: []Entry{
				{File: "main.go", Message: "error1", ContextHash: "hash1", Count: 1},
			},
			wantLen:      1,
			wantCountFor: map[string]int{"main.go:error1:hash1": 1},
		},
		{
			name: "duplicate entry increments count",
			entries: []Entry{
				{File: "main.go", Message: "error1", ContextHash: "hash1", Count: 1},
				{File: "main.go", Message: "error1", ContextHash: "hash1", Count: 1},
				{File: "main.go", Message: "error1", ContextHash: "hash1", Count: 1},
			},
			wantLen:      1,
			wantCountFor: map[string]int{"main.go:error1:hash1": 3},
		},
		{
			name: "different files",
			entries: []Entry{
				{File: "a.go", Message: "error", ContextHash: "h1", Count: 1},
				{File: "b.go", Message: "error", ContextHash: "h2", Count: 1},
			},
			wantLen: 2,
			wantCountFor: map[string]int{
				"a.go:error:h1": 1,
				"b.go:error:h2": 1,
			},
		},
		{
			name: "same file different messages",
			entries: []Entry{
				{File: "main.go", Message: "error1", ContextHash: "h1", Count: 1},
				{File: "main.go", Message: "error2", ContextHash: "h2", Count: 1},
			},
			wantLen: 2,
			wantCountFor: map[string]int{
				"main.go:error1:h1": 1,
				"main.go:error2:h2": 1,
			},
		},
		{
			name: "same file and message different hash",
			entries: []Entry{
				{File: "main.go", Message: "error", ContextHash: "hash1", Count: 1},
				{File: "main.go", Message: "error", ContextHash: "hash2", Count: 1},
			},
			wantLen: 2,
			wantCountFor: map[string]int{
				"main.go:error:hash1": 1,
				"main.go:error:hash2": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := New()

			for _, e := range tt.entries {
				bl.Add(e)
			}

			if bl.Len() != tt.wantLen {
				t.Errorf("Len() = %d, want %d", bl.Len(), tt.wantLen)
			}

			for key, wantCount := range tt.wantCountFor {
				found := false
				for _, e := range bl.Entries {
					entryKey := e.File + ":" + e.Message + ":" + e.ContextHash
					if entryKey == key {
						if e.Count != wantCount {
							t.Errorf("Count for %q = %d, want %d", key, e.Count, wantCount)
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Entry %q not found", key)
				}
			}
		})
	}
}

func TestBaseline_FindByFileAndMessage(t *testing.T) {
	bl := New()
	bl.Add(Entry{File: "main.go", Message: "error1", ContextHash: "h1", Count: 1})
	bl.Add(Entry{File: "main.go", Message: "error1", ContextHash: "h2", Count: 1})
	bl.Add(Entry{File: "main.go", Message: "error2", ContextHash: "h3", Count: 1})
	bl.Add(Entry{File: "other.go", Message: "error1", ContextHash: "h4", Count: 1})

	tests := []struct {
		name        string
		file        string
		message     string
		wantMatches int
	}{
		{
			name:        "multiple matches same file and message",
			file:        "main.go",
			message:     "error1",
			wantMatches: 2,
		},
		{
			name:        "single match",
			file:        "main.go",
			message:     "error2",
			wantMatches: 1,
		},
		{
			name:        "different file same message",
			file:        "other.go",
			message:     "error1",
			wantMatches: 1,
		},
		{
			name:        "no match - wrong file",
			file:        "notexist.go",
			message:     "error1",
			wantMatches: 0,
		},
		{
			name:        "no match - wrong message",
			file:        "main.go",
			message:     "nonexistent",
			wantMatches: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := bl.FindByFileAndMessage(tt.file, tt.message)
			if len(matches) != tt.wantMatches {
				t.Errorf("FindByFileAndMessage() returned %d matches, want %d", len(matches), tt.wantMatches)
			}
		})
	}
}

func TestBaseline_Len(t *testing.T) {
	bl := New()

	if bl.Len() != 0 {
		t.Errorf("Len() on empty baseline = %d, want 0", bl.Len())
	}

	bl.Add(Entry{File: "a.go", Message: "msg", ContextHash: "h1", Count: 1})
	if bl.Len() != 1 {
		t.Errorf("Len() after 1 add = %d, want 1", bl.Len())
	}

	bl.Add(Entry{File: "b.go", Message: "msg", ContextHash: "h2", Count: 1})
	if bl.Len() != 2 {
		t.Errorf("Len() after 2 adds = %d, want 2", bl.Len())
	}

	// Adding duplicate should not increase length
	bl.Add(Entry{File: "a.go", Message: "msg", ContextHash: "h1", Count: 1})
	if bl.Len() != 2 {
		t.Errorf("Len() after duplicate add = %d, want 2", bl.Len())
	}
}
