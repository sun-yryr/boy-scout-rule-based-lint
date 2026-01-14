package baseline

// Matcher handles matching lint issues against the baseline
type Matcher struct{}

// NewMatcher creates a new Matcher
func NewMatcher() *Matcher {
	return &Matcher{}
}

// Match checks if an entry matches any entry in the baseline
// It first matches by file and message, then by context hash
func (m *Matcher) Match(bl *Baseline, entry Entry) bool {
	candidates := bl.FindByFileAndMessage(entry.File, entry.Message)
	if len(candidates) == 0 {
		return false
	}

	// Check for exact context hash match
	for _, c := range candidates {
		if c.ContextHash == entry.ContextHash {
			return true
		}
	}

	// If no exact match and context hash is empty, match by file+message only
	if entry.ContextHash == "" {
		return true
	}

	// Fallback: if the issue's context couldn't be extracted but
	// there are baseline entries for the same file+message,
	// we can optionally be lenient
	for _, c := range candidates {
		if c.ContextHash == "" {
			return true
		}
	}

	return false
}
