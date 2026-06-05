package context

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// Context represents the code context around a lint issue
type Context struct {
	Lines []string
	Hash  string
}

// Extractor extracts code context from source files
type Extractor struct{}

// NewExtractor creates a new Extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract extracts the context around a specific line in a file
func (e *Extractor) Extract(filePath string, lineNum int) (ctx *Context, err error) {
	if lineNum <= 0 {
		return nil, fmt.Errorf("invalid line number %d: must be positive", lineNum)
	}

	var contextLines []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	scanner := bufio.NewScanner(file)
	currentLine := 0
	found := false
	for scanner.Scan() {
		currentLine++
		if currentLine == lineNum {
			contextLines = append(contextLines, scanner.Text())
			found = true
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("line %d not found in %s", lineNum, filePath)
	}

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
