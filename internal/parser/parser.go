package parser

// Issue represents a single lint issue
type Issue struct {
	File    string
	Line    int
	Column  int
	Message string
	Raw     string
}

// Parser is an interface for parsing lint output
type Parser interface {
	Parse(line string) (*Issue, error)
}
