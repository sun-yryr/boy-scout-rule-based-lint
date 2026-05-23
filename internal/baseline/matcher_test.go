package baseline

import (
	"testing"
)

func TestMatcher_Match(t *testing.T) {
	tests := []struct {
		name    string
		bl      *Baseline
		entries []Entry // entries to match sequentially, each with expected result
		want    []bool
	}{
		{
			name: "single match within count",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 1})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "error", ContextHash: "h1"},
			},
			want: []bool{true},
		},
		{
			name: "match count exceeded",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 1})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "error", ContextHash: "h1"},
				{File: "a.go", Message: "error", ContextHash: "h1"},
			},
			want: []bool{true, false},
		},
		{
			name: "count 3 suppresses first 3 then reports",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 3})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "error", ContextHash: "h1"},
				{File: "a.go", Message: "error", ContextHash: "h1"},
				{File: "a.go", Message: "error", ContextHash: "h1"},
				{File: "a.go", Message: "error", ContextHash: "h1"},
			},
			want: []bool{true, true, true, false},
		},
		{
			name: "different entries tracked separately",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "err1", ContextHash: "h1", Count: 1})
				bl.Add(Entry{File: "b.go", Message: "err2", ContextHash: "h2", Count: 2})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "err1", ContextHash: "h1"},
				{File: "a.go", Message: "err1", ContextHash: "h1"},
				{File: "b.go", Message: "err2", ContextHash: "h2"},
				{File: "b.go", Message: "err2", ContextHash: "h2"},
				{File: "b.go", Message: "err2", ContextHash: "h2"},
			},
			want: []bool{true, false, true, true, false},
		},
		{
			name: "no match - different file",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 1})
				return bl
			}(),
			entries: []Entry{
				{File: "b.go", Message: "error", ContextHash: "h1"},
			},
			want: []bool{false},
		},
		{
			name: "no match - different message",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 1})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "other", ContextHash: "h1"},
			},
			want: []bool{false},
		},
		{
			name: "empty context hash matches any candidate",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 2})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "error", ContextHash: ""},
				{File: "a.go", Message: "error", ContextHash: ""},
				{File: "a.go", Message: "error", ContextHash: ""},
			},
			want: []bool{true, true, false},
		},
		{
			name: "baseline entry with empty hash acts as fallback",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "", Count: 1})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "error", ContextHash: "h1"},
				{File: "a.go", Message: "error", ContextHash: "h1"},
			},
			want: []bool{true, false},
		},
		{
			name: "exact match exhausted does not fall through to empty hash fallback",
			bl: func() *Baseline {
				bl := New()
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 1})
				bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "", Count: 5})
				return bl
			}(),
			entries: []Entry{
				{File: "a.go", Message: "error", ContextHash: "h1"},
				{File: "a.go", Message: "error", ContextHash: "h1"},
			},
			want: []bool{true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher()
			for i, entry := range tt.entries {
				got := m.Match(tt.bl, entry)
				if got != tt.want[i] {
					t.Errorf("Match() at index %d = %v, want %v (entry: %+v)", i, got, tt.want[i], entry)
				}
			}
		})
	}
}

func TestMatcher_NewInstanceResetsCount(t *testing.T) {
	bl := New()
	bl.Add(Entry{File: "a.go", Message: "error", ContextHash: "h1", Count: 1})

	m1 := NewMatcher()
	if !m1.Match(bl, Entry{File: "a.go", Message: "error", ContextHash: "h1"}) {
		t.Error("first matcher, first match should be true")
	}
	if m1.Match(bl, Entry{File: "a.go", Message: "error", ContextHash: "h1"}) {
		t.Error("first matcher, second match should be false")
	}

	m2 := NewMatcher()
	if !m2.Match(bl, Entry{File: "a.go", Message: "error", ContextHash: "h1"}) {
		t.Error("second matcher, first match should be true (fresh state)")
	}
}
