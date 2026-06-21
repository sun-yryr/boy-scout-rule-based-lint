package baseline

import (
	"bytes"
	"encoding/json"
	"os"
)

// Store handles reading and writing baseline files
type Store struct{}

// NewStore creates a new Store
func NewStore() *Store {
	return &Store{}
}

// Load loads a baseline from a file
func (s *Store) Load(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bl Baseline
	if err := json.Unmarshal(data, &bl); err != nil {
		return nil, err
	}

	return &bl, nil
}

// Save saves a baseline to a file
func (s *Store) Save(path string, bl *Baseline) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(bl); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}
