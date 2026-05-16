package scope

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/scope/treesitter"
)

type tag struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Line  int    `json:"line"`
	End   int    `json:"end"`
	Scope string `json:"scope"`
}

type Tagger interface {
	Scope(filePath string, lineNum int) (string, error)
}

type CtagsTagger struct{}

func NewCtagsTagger() *CtagsTagger {
	return &CtagsTagger{}
}

type TreesitterTagger struct{}

func NewTreesitterTagger() *TreesitterTagger {
	return &TreesitterTagger{}
}

func (t *TreesitterTagger) Scope(filePath string, lineNum int) (string, error) {
	lang := languageFromExt(filePath)
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")

	source, err := os.ReadFile(filePath)
	if err != nil {
		return fileScope(filePath), nil
	}

	scopeNode, err := treesitter.Analyze(source, lineNum, ext)
	if err != nil || scopeNode == nil {
		return fileScope(filePath), nil
	}

	return fmt.Sprintf("%s:%s:%s", lang, scopeNode.Kind, treesitter.FormatScope(scopeNode)), nil
}

func (t *CtagsTagger) Scope(filePath string, lineNum int) (string, error) {
	if !ctagsAvailable() {
		return fileScope(filePath), nil
	}

	abs, err := filepath.Abs(filePath)
	if err != nil {
		abs = filePath
	}

	cmd := exec.Command("ctags", "--output-format=json", "--fields=+ne", abs)
	output, err := cmd.Output()
	if err != nil {
		return fileScope(filePath), nil
	}

	var tags []tag
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		var t tag
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			continue
		}
		if !isScopeKind(t.Kind) {
			continue
		}
		tags = append(tags, t)
	}

	enclosing := findEnclosing(tags, lineNum)
	if enclosing == nil {
		return fileScope(filePath), nil
	}

	return formatScope(filePath, enclosing), nil
}

func ctagsAvailable() bool {
	_, err := exec.LookPath("ctags")
	return err == nil
}

func isScopeKind(kind string) bool {
	switch kind {
	case "func", "function", "method",
		"class", "struct", "interface",
		"namespace", "module", "type", "enum", "trait":
		return true
	}
	return false
}

func findEnclosing(tags []tag, lineNum int) *tag {
	var best *tag
	for i := range tags {
		t := &tags[i]

		if t.Line > lineNum {
			continue
		}

		if t.End > 0 && t.End < lineNum {
			continue
		}

		if best == nil || t.Line > best.Line {
			best = t
		}
	}
	return best
}

func formatScope(filePath string, tag *tag) string {
	lang := languageFromExt(filePath)
	kind := normalizeKind(tag.Kind)

	name := tag.Name
	if tag.Scope != "" {
		name = tag.Scope + "." + name
	}

	return fmt.Sprintf("%s:%s:%s", lang, kind, name)
}

func fileScope(filePath string) string {
	lang := languageFromExt(filePath)
	return fmt.Sprintf("%s:%s", lang, "file")
}

func normalizeKind(kind string) string {
	switch kind {
	case "function":
		return "func"
	case "member", "field":
		return "field"
	default:
		return kind
	}
}

func languageFromExt(filePath string) string {
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
	switch ext {
	case "go":
		return "go"
	case "py":
		return "python"
	case "js":
		return "javascript"
	case "ts":
		return "typescript"
	case "tsx":
		return "typescript"
	case "jsx":
		return "javascript"
	case "php":
		return "php"
	case "java":
		return "java"
	case "rs":
		return "rust"
	case "rb":
		return "ruby"
	case "c":
		return "c"
	case "cpp", "cc", "cxx":
		return "cpp"
	case "h", "hpp":
		return "cpp"
	default:
		return ext
	}
}
