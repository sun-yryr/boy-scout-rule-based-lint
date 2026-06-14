package baseline

import (
	"fmt"
	"testing"
)

func entry(file, msg, hash string, count int) Entry {
	return Entry{
		File:    file,
		Message: msg,
		Count:   count,
		Fingerprints: Fingerprints{
			LineHash: hash,
		},
	}
}

func TestExactMatcher(t *testing.T) {
	m := NewExactMatcher()

	tests := []struct {
		name      string
		base      Entry
		current   Entry
		wantKey   string
		wantMatch bool
	}{
		{
			name:      "same file and hash match",
			base:      entry("main.go", "err", "abc", 1),
			current:   entry("main.go", "err", "abc", 1),
			wantKey:   "main.go:abc",
			wantMatch: true,
		},
		{
			name:      "same file different hash",
			base:      entry("main.go", "err", "abc", 1),
			current:   entry("main.go", "err", "def", 1),
			wantKey:   "",
			wantMatch: false,
		},
		{
			name:      "different file same hash",
			base:      entry("main.go", "err", "abc", 1),
			current:   entry("other.go", "err", "abc", 1),
			wantKey:   "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, key := m.Match(tt.base, tt.current)
			if matched != tt.wantMatch {
				t.Errorf("Match() = %v, want %v", matched, tt.wantMatch)
			}
			if key != tt.wantKey {
				t.Errorf("Match() key = %q, want %q", key, tt.wantKey)
			}
		})
	}
}

func TestExactMatcher_GroupKey(t *testing.T) {
	m := NewExactMatcher()
	e := entry("main.go", "unused var", "abc123", 1)
	got := m.GroupKey(e)
	want := "main.go:abc123"
	if got != want {
		t.Errorf("GroupKey() = %q, want %q", got, want)
	}
}

func TestLooseMatcher(t *testing.T) {
	m := NewLooseMatcher()

	tests := []struct {
		name      string
		base      Entry
		current   Entry
		wantKey   string
		wantMatch bool
	}{
		{
			name:      "same file and message match",
			base:      entry("main.go", "unused var", "abc", 1),
			current:   entry("main.go", "unused var", "def", 1),
			wantKey:   "main.go:unused var",
			wantMatch: true,
		},
		{
			name:      "same file different message",
			base:      entry("main.go", "unused var", "abc", 1),
			current:   entry("main.go", "other err", "abc", 1),
			wantKey:   "",
			wantMatch: false,
		},
		{
			name:      "different file same message",
			base:      entry("main.go", "unused var", "abc", 1),
			current:   entry("other.go", "unused var", "abc", 1),
			wantKey:   "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, key := m.Match(tt.base, tt.current)
			if matched != tt.wantMatch {
				t.Errorf("Match() = %v, want %v", matched, tt.wantMatch)
			}
			if key != tt.wantKey {
				t.Errorf("Match() key = %q, want %q", key, tt.wantKey)
			}
		})
	}
}

func TestLooseMatcher_GroupKey(t *testing.T) {
	m := NewLooseMatcher()
	e := entry("main.go", "unused var", "abc123", 1)
	got := m.GroupKey(e)
	want := "main.go:unused var"
	if got != want {
		t.Errorf("GroupKey() = %q, want %q", got, want)
	}
}

func TestSessionMatcher_Exact(t *testing.T) {
	bl := New()
	bl.Add(entry("main.go", "err1", "hash1", 1))
	bl.Add(entry("main.go", "err2", "hash2", 1))
	bl.Add(entry("other.go", "err1", "hash3", 1))

	sm := NewSessionMatcher(bl, NewExactMatcher())

	tests := []struct {
		name    string
		current Entry
		want    bool
	}{
		{"exact match", entry("main.go", "err1", "hash1", 1), true},
		{"no match different hash", entry("main.go", "err1", "hashX", 1), false},
		{"second exact match", entry("main.go", "err2", "hash2", 1), true},
		{"third exact match", entry("other.go", "err1", "hash3", 1), true},
		{"re-consume already matched", entry("main.go", "err1", "hash1", 1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sm.Match(tt.current)
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionMatcher_Loose(t *testing.T) {
	bl := New()
	bl.Add(entry("main.go", "unused var", "hash_a", 1))
	bl.Add(entry("main.go", "unused var", "hash_b", 1))
	bl.Add(entry("main.go", "unused var", "hash_c", 1))

	sm := NewSessionMatcher(bl, NewLooseMatcher())

	current := entry("main.go", "unused var", "hash_current", 1)

	if !sm.Match(current) {
		t.Error("Match() 1st = false, want true")
	}
	if !sm.Match(current) {
		t.Error("Match() 2nd = false, want true")
	}
	if !sm.Match(current) {
		t.Error("Match() 3rd = false, want true (count=3)")
	}
	if sm.Match(current) {
		t.Error("Match() 4th = true, want false (count exhausted)")
	}
}

func TestSessionMatcher_LooseMixedFiles(t *testing.T) {
	bl := New()
	bl.Add(entry("main.go", "unused", "h1", 1))
	bl.Add(entry("main.go", "unused", "h2", 1))
	bl.Add(entry("lib.go", "unused", "h3", 1))

	sm := NewSessionMatcher(bl, NewLooseMatcher())

	// main.go:unused の remaining は 2
	if !sm.Match(entry("main.go", "unused", "hx", 1)) {
		t.Error("main.go:unused 1st should match")
	}
	if !sm.Match(entry("main.go", "unused", "hy", 1)) {
		t.Error("main.go:unused 2nd should match")
	}
	if sm.Match(entry("main.go", "unused", "hz", 1)) {
		t.Error("main.go:unused 3rd should NOT match (count exhausted)")
	}

	// lib.go:unused の remaining は 1
	if !sm.Match(entry("lib.go", "unused", "hw", 1)) {
		t.Error("lib.go:unused 1st should match")
	}
	if sm.Match(entry("lib.go", "unused", "hv", 1)) {
		t.Error("lib.go:unused 2nd should NOT match (count exhausted)")
	}

	// 関係ないファイル
	if sm.Match(entry("other.go", "unused", "hu", 1)) {
		t.Error("other.go:unused should NOT match")
	}
}

func TestExactMatcher_MatchAny(t *testing.T) {
	m := NewExactMatcher()
	base := entry("main.go", "err", "hash1", 1)

	candidates := []Entry{
		entry("main.go", "err", "hashX", 1),
		entry("main.go", "err", "hash1", 1),
		entry("main.go", "err", "hashY", 1),
	}

	matched, key := m.MatchAny(base, candidates)
	if !matched {
		t.Error("MatchAny() should match")
	}
	if key != "main.go:hash1" {
		t.Errorf("MatchAny() key = %q, want %q", key, "main.go:hash1")
	}
}

func TestExactMatcher_MatchAnyNoMatch(t *testing.T) {
	m := NewExactMatcher()
	base := entry("main.go", "err", "hash1", 1)

	candidates := []Entry{
		entry("main.go", "err", "hashX", 1),
		entry("main.go", "err", "hashY", 1),
	}

	matched, key := m.MatchAny(base, candidates)
	if matched {
		t.Error("MatchAny() should not match")
	}
	if key != "" {
		t.Errorf("MatchAny() key = %q, want empty", key)
	}
}

func TestLooseMatcher_MatchAny(t *testing.T) {
	m := NewLooseMatcher()
	base := entry("main.go", "unused var", "h1", 1)

	candidates := []Entry{
		entry("main.go", "other err", "hX", 1),
		entry("main.go", "unused var", "h2", 1),
	}

	matched, key := m.MatchAny(base, candidates)
	if !matched {
		t.Error("MatchAny() should match")
	}
	if key != "main.go:unused var" {
		t.Errorf("MatchAny() key = %q, want %q", key, "main.go:unused var")
	}
}

func TestLooseMatcher_MatchAnyNoMatch(t *testing.T) {
	m := NewLooseMatcher()
	base := entry("main.go", "unused var", "h1", 1)

	candidates := []Entry{
		entry("main.go", "other err", "hX", 1),
	}

	matched, key := m.MatchAny(base, candidates)
	if matched {
		t.Error("MatchAny() should not match")
	}
	if key != "" {
		t.Errorf("MatchAny() key = %q, want empty", key)
	}
}

func TestSessionMatcher_CountSummation(t *testing.T) {
	// Count>1 のエントリをきちんと消費できるか
	bl := New()
	bl.Add(Entry{
		File:    "main.go",
		Message: "unused",
		Count:   3,
		Fingerprints: Fingerprints{
			LineHash: "abc",
		},
	})

	sm := NewSessionMatcher(bl, NewLooseMatcher())
	current := entry("main.go", "unused", "hx", 1)

	for i := 0; i < 3; i++ {
		if !sm.Match(current) {
			t.Errorf("Match() %d = false, want true", i+1)
		}
	}
	if sm.Match(current) {
		t.Error("Match() 4th = true, want false (count=3 exhausted)")
	}
}

// Entry with zero Count should not add to remaining
func TestSessionMatcher_ZeroCount(t *testing.T) {
	t.Skip("zero count entries are invalid; not implemented yet")
}

// Verify remaining state after partial consumption
func TestSessionMatcher_RemainingState(t *testing.T) {
	bl := New()
	bl.Add(entry("a.go", "err1", "h1", 1))
	bl.Add(entry("a.go", "err1", "h2", 1))
	bl.Add(entry("b.go", "err1", "h3", 1))

	sm := NewSessionMatcher(bl, NewLooseMatcher())

	if sm.remaining["a.go:err1"] != 2 {
		t.Errorf("initial remaining[a.go:err1] = %d, want 2", sm.remaining["a.go:err1"])
	}
	if sm.remaining["b.go:err1"] != 1 {
		t.Errorf("initial remaining[b.go:err1] = %d, want 1", sm.remaining["b.go:err1"])
	}

	sm.Match(entry("a.go", "err1", "hx", 1))

	if sm.remaining["a.go:err1"] != 1 {
		t.Errorf("remaining[a.go:err1] after 1 consume = %d, want 1", sm.remaining["a.go:err1"])
	}
}

func TestExactMatcher_GroupKeyEmptyHash(t *testing.T) {
	m := NewExactMatcher()
	e := entry("main.go", "msg", "", 1)
	key := m.GroupKey(e)
	if key != "main.go:" {
		t.Errorf("GroupKey with empty hash = %q, want %q", key, "main.go:")
	}
}

func TestLooseMatcher_GroupKeyEmptyMessage(t *testing.T) {
	m := NewLooseMatcher()
	e := entry("main.go", "", "abc", 1)
	key := m.GroupKey(e)
	if key != "main.go:" {
		t.Errorf("GroupKey with empty message = %q, want %q", key, "main.go:")
	}
}

type offsetMatcher struct {
	MatchStrategy
}

func (m *offsetMatcher) GroupKey(e Entry) string {
	return fmt.Sprintf("%s:%s", e.File, e.Fingerprints.LineHash)
}

func (m *offsetMatcher) Match(base, current Entry) (bool, string) {
	key := m.GroupKey(base)
	if m.GroupKey(current) == key {
		return true, key
	}
	return false, ""
}

func (m *offsetMatcher) MatchAny(base Entry, candidates []Entry) (bool, string) {
	key := m.GroupKey(base)
	for i := range candidates {
		if m.GroupKey(candidates[i]) == key {
			return true, key
		}
	}
	return false, ""
}

// Custom strategy using LineHash as GroupKey, simulating smart/scope style
// where GroupKey is a unique identifier not used for match.
func TestSessionMatcher_CustomStrategyUniqueKeys(t *testing.T) {
	bl := New()
	bl.Add(Entry{File: "main.go", Message: "err", Count: 1, Fingerprints: Fingerprints{LineHash: "h1"}})
	bl.Add(Entry{File: "main.go", Message: "err", Count: 1, Fingerprints: Fingerprints{LineHash: "h2"}})

	sm := NewSessionMatcher(bl, &offsetMatcher{})

	if sm.remaining["main.go:h1"] != 1 {
		t.Errorf("remaining[main.go:h1] = %d, want 1", sm.remaining["main.go:h1"])
	}
	if sm.remaining["main.go:h2"] != 1 {
		t.Errorf("remaining[main.go:h2] = %d, want 1", sm.remaining["main.go:h2"])
	}

	if !sm.Match(Entry{File: "main.go", Message: "err", Count: 1, Fingerprints: Fingerprints{LineHash: "h1"}}) {
		t.Error("should match first entry")
	}
	if sm.Match(Entry{File: "main.go", Message: "err", Count: 1, Fingerprints: Fingerprints{LineHash: "h1"}}) {
		t.Error("should not match first entry again (count exhausted)")
	}

	if !sm.Match(Entry{File: "main.go", Message: "err", Count: 1, Fingerprints: Fingerprints{LineHash: "h2"}}) {
		t.Error("should match second entry")
	}
}
