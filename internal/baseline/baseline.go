package baseline

// Entry represents a single baseline entry
type Entry struct {
	File         string   `json:"file"`
	Message      string   `json:"message"`
	ContextHash  string   `json:"context_hash"`
	ContextLines []string `json:"context_lines"`
	Count        int      `json:"count"`
}

// Baseline represents a collection of baseline entries
type Baseline struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

// New creates a new empty Baseline
func New() *Baseline {
	return &Baseline{
		Version: 1,
		Entries: []Entry{},
	}
}

// Add adds an entry to the baseline
// If a matching entry exists (same file, message, and context hash), increment the count
func (b *Baseline) Add(entry Entry) {
	for i, e := range b.Entries {
		if e.File == entry.File && e.Message == entry.Message && e.ContextHash == entry.ContextHash {
			b.Entries[i].Count++
			return
		}
	}
	b.Entries = append(b.Entries, entry)
}

// Len returns the number of entries in the baseline
func (b *Baseline) Len() int {
	return len(b.Entries)
}

// FindByFileAndMessage returns all entries matching the given file and message
func (b *Baseline) FindByFileAndMessage(file, message string) []Entry {
	var matches []Entry
	for _, e := range b.Entries {
		if e.File == file && e.Message == message {
			matches = append(matches, e)
		}
	}
	return matches
}
