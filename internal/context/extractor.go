package context

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
)

// Context represents the code context around a lint issue
type Context struct {
	Lines []string
	Hash  string
}

// Extractor extracts code context from source files
type Extractor struct {
	contextLines int
}

// NewExtractor creates a new Extractor
func NewExtractor(contextLines int) *Extractor {
	return &Extractor{
		contextLines: contextLines,
	}
}

// Extract extracts the context around a specific line in a file
func (e *Extractor) Extract(filePath string, lineNum int) (*Context, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	currentLine := 0
	for scanner.Scan() {
		currentLine++
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Calculate start and end indices (0-indexed)
	startIdx := lineNum - 1 - e.contextLines
	endIdx := lineNum - 1 + e.contextLines

	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx >= len(lines) {
		endIdx = len(lines) - 1
	}

	// Extract context lines
	contextLines := make([]string, 0, endIdx-startIdx+1)
	for i := startIdx; i <= endIdx; i++ {
		contextLines = append(contextLines, lines[i])
	}

	// Compute hash
	hash := computeHash(contextLines)

	return &Context{
		Lines: contextLines,
		Hash:  hash,
	}, nil
}

// computeHash computes a hash of the context lines
// It normalizes whitespace before hashing
func computeHash(lines []string) string {
	var normalized []string
	for _, line := range lines {
		// Normalize whitespace: trim and collapse multiple spaces
		trimmed := strings.TrimSpace(line)
		normalized = append(normalized, trimmed)
	}

	combined := strings.Join(normalized, "\n")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}
