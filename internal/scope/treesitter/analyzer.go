package treesitter

import (
	"fmt"
	"unsafe"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type ScopeNode struct {
	Kind   string
	Name   string
	Parent string
}

type LanguageAnalyzer interface {
	Language() unsafe.Pointer
	FindScope(node *sitter.Node, source []byte) (*ScopeNode, error)
}

type analyzerEntry struct {
	exts     []string
	analyzer LanguageAnalyzer
}

var registry []analyzerEntry

func Register(analyzer LanguageAnalyzer, exts ...string) {
	registry = append(registry, analyzerEntry{
		exts:     exts,
		analyzer: analyzer,
	})
}

func FindAnalyzer(ext string) LanguageAnalyzer {
	for _, entry := range registry {
		for _, e := range entry.exts {
			if e == ext {
				return entry.analyzer
			}
		}
	}
	return nil
}

func findDeepestNode(root *sitter.Node, lineNum int) *sitter.Node {
	row := uint(lineNum - 1)
	point := sitter.NewPoint(row, 0)
	return root.NamedDescendantForPointRange(point, point)
}

func nodeName(node *sitter.Node, source []byte) string {
	name := node.ChildByFieldName("name")
	if name == nil {
		return ""
	}
	return name.Utf8Text(source)
}

func FormatScope(node *ScopeNode) string {
	if node.Parent != "" {
		return fmt.Sprintf("%s.%s", node.Parent, node.Name)
	}
	return node.Name
}

func Analyze(source []byte, lineNum int, ext string) (*ScopeNode, error) {
	analyzer := FindAnalyzer(ext)
	if analyzer == nil {
		return nil, fmt.Errorf("no analyzer for extension %q", ext)
	}

	parser := sitter.NewParser()
	defer parser.Close()

	language := sitter.NewLanguage(analyzer.Language())
	if err := parser.SetLanguage(language); err != nil {
		return nil, fmt.Errorf("set language: %w", err)
	}

	tree := parser.Parse(source, nil)
	defer tree.Close()

	node := findDeepestNode(tree.RootNode(), lineNum)
	return analyzer.FindScope(node, source)
}
