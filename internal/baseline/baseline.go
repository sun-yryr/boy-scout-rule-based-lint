package baseline

type Fingerprints struct {
	LineHash string `json:"line_hash"`
}

// Config holds optional baseline-level settings.
type Config struct {
	BoyScoutPolicy string `json:"boy_scout_policy,omitempty"`
	BaseRef        string `json:"base_ref,omitempty"`
}

// Entry represents a single baseline entry
type Entry struct {
	File         string       `json:"file"`
	Message      string       `json:"message"`
	SourceLine   string       `json:"source_line"`
	Count        int          `json:"count"`
	Fingerprints Fingerprints `json:"fingerprints"`
}

// Baseline represents a collection of baseline entries
type Baseline struct {
	Version int     `json:"version"`
	Config  *Config `json:"config,omitempty"`
	Entries []Entry `json:"entries"`
}

// New creates a new empty Baseline
func New() *Baseline {
	return &Baseline{
		Version: 2,
		Entries: []Entry{},
	}
}

// Add appends an entry to the baseline without checking for duplicates
func (b *Baseline) Add(entry Entry) {
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
