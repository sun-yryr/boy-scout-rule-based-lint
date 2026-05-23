package baseline

// Matcher handles matching lint issues against the baseline.
// It tracks how many times each baseline entry has been matched,
// so that if the match count exceeds entry.Count, the issue is
// no longer suppressed (i.e. the excess is reported as a new issue).
type Matcher struct {
	matchedCount map[string]int // key: "file:message:hash" -> matches consumed so far
}

// NewMatcher creates a new Matcher
func NewMatcher() *Matcher {
	return &Matcher{
		matchedCount: make(map[string]int),
	}
}

// Match checks if an entry matches any entry in the baseline
// It first matches by file and message, then by context hash.
// Returns true only if the matched baseline entry still has
// remaining count to consume.
func (m *Matcher) Match(bl *Baseline, entry Entry) bool {
	candidates := bl.FindByFileAndMessage(entry.File, entry.Message)
	if len(candidates) == 0 {
		return false
	}

	// Check for exact context hash match
	for _, c := range candidates {
		if c.ContextHash == entry.ContextHash {
			key := c.File + ":" + c.Message + ":" + c.ContextHash
			if m.matchedCount[key] < c.Count {
				m.matchedCount[key]++
				return true
			}
			return false
		}
	}

	// If no exact match and context hash is empty, match any candidate
	if entry.ContextHash == "" {
		for _, c := range candidates {
			key := c.File + ":" + c.Message + ":" + c.ContextHash
			if m.matchedCount[key] < c.Count {
				m.matchedCount[key]++
				return true
			}
		}
		return false
	}

	// Fallback: baseline entries with empty hash
	for _, c := range candidates {
		if c.ContextHash == "" {
			key := c.File + ":" + c.Message + ":"
			if m.matchedCount[key] < c.Count {
				m.matchedCount[key]++
				return true
			}
		}
	}

	return false
}
