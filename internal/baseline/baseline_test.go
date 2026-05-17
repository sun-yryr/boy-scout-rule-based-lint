package baseline

import (
	"testing"
)

func TestNew(t *testing.T) {
	bl := New()

	if bl.Version != 2 {
		t.Errorf("Version = %d, want 2", bl.Version)
	}
	if bl.Len() != 0 {
		t.Errorf("Len() = %d, want 0", bl.Len())
	}
}

func TestBaseline_Add(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		wantLen int
	}{
		{
			name: "single entry",
			entries: []Entry{
				{File: "main.go", Message: "error1", Count: 1},
			},
			wantLen: 1,
		},
		{
			name: "duplicate entry does not deduplicate",
			entries: []Entry{
				{File: "main.go", Message: "error1", Count: 1},
				{File: "main.go", Message: "error1", Count: 1},
				{File: "main.go", Message: "error1", Count: 1},
			},
			wantLen: 3,
		},
		{
			name: "different files",
			entries: []Entry{
				{File: "a.go", Message: "error", Count: 1},
				{File: "b.go", Message: "error", Count: 1},
			},
			wantLen: 2,
		},
		{
			name: "same file different messages",
			entries: []Entry{
				{File: "main.go", Message: "error1", Count: 1},
				{File: "main.go", Message: "error2", Count: 1},
			},
			wantLen: 2,
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
		})
	}
}

func TestBaseline_FindByFileAndMessage(t *testing.T) {
	bl := New()
	bl.Add(Entry{File: "main.go", Message: "error1", Count: 1})
	bl.Add(Entry{File: "main.go", Message: "error1", Count: 1})
	bl.Add(Entry{File: "main.go", Message: "error2", Count: 1})
	bl.Add(Entry{File: "other.go", Message: "error1", Count: 1})

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

	bl.Add(Entry{File: "a.go", Message: "msg", Count: 1})
	if bl.Len() != 1 {
		t.Errorf("Len() after 1 add = %d, want 1", bl.Len())
	}

	bl.Add(Entry{File: "b.go", Message: "msg", Count: 1})
	if bl.Len() != 2 {
		t.Errorf("Len() after 2 adds = %d, want 2", bl.Len())
	}

	bl.Add(Entry{File: "a.go", Message: "msg", Count: 1})
	if bl.Len() != 3 {
		t.Errorf("Len() after 3 adds = %d, want 3", bl.Len())
	}
}
