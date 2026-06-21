package baseline

type MatchStrategy interface {
	GroupKey(e Entry) string
	Match(base, current Entry) (matched bool, key string)
	MatchAny(base Entry, candidates []Entry) (matched bool, key string)
}

type SessionMatcher struct {
	baseline  *Baseline
	strategy  MatchStrategy
	remaining map[string]int
}

func NewSessionMatcher(bl *Baseline, s MatchStrategy) *SessionMatcher {
	sm := &SessionMatcher{
		baseline:  bl,
		strategy:  s,
		remaining: make(map[string]int, len(bl.Entries)),
	}
	for _, e := range bl.Entries {
		sm.remaining[s.GroupKey(e)] += e.Count
	}
	return sm
}

func (sm *SessionMatcher) Match(current Entry) bool {
	for i := range sm.baseline.Entries {
		base := sm.baseline.Entries[i]
		matched, key := sm.strategy.Match(base, current)
		if !matched {
			continue
		}
		if sm.remaining[key] <= 0 {
			continue
		}
		sm.remaining[key]--
		return true
	}
	return false
}

type ExactMatcher struct{}

func NewExactMatcher() *ExactMatcher {
	return &ExactMatcher{}
}

func (m *ExactMatcher) GroupKey(e Entry) string {
	return e.File + ":" + e.Fingerprints.LineHash
}

func (m *ExactMatcher) Match(base, current Entry) (bool, string) {
	key := m.GroupKey(base)
	if m.GroupKey(current) == key {
		return true, key
	}
	return false, ""
}

func (m *ExactMatcher) MatchAny(base Entry, candidates []Entry) (bool, string) {
	key := m.GroupKey(base)
	for i := range candidates {
		if m.GroupKey(candidates[i]) == key {
			return true, key
		}
	}
	return false, ""
}

type LooseMatcher struct{}

func NewLooseMatcher() *LooseMatcher {
	return &LooseMatcher{}
}

func (m *LooseMatcher) GroupKey(e Entry) string {
	return e.File + ":" + e.Message
}

func (m *LooseMatcher) Match(base, current Entry) (bool, string) {
	key := m.GroupKey(base)
	if m.GroupKey(current) == key {
		return true, key
	}
	return false, ""
}

func (m *LooseMatcher) MatchAny(base Entry, candidates []Entry) (bool, string) {
	key := m.GroupKey(base)
	for i := range candidates {
		if m.GroupKey(candidates[i]) == key {
			return true, key
		}
	}
	return false, ""
}
