package treesitter

import (
	"unsafe"

	sitter "github.com/tree-sitter/go-tree-sitter"
	ts "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

type typescriptAnalyzer struct{}

func (a *typescriptAnalyzer) Language() unsafe.Pointer {
	return ts.LanguageTypescript()
}

func (a *typescriptAnalyzer) FindScope(node *sitter.Node, source []byte) (*ScopeNode, error) {
	if node == nil {
		return nil, nil
	}

	for n := node; n != nil; n = n.Parent() {
		switch n.Kind() {
		case "function_declaration":
			name := nodeName(n, source)
			if name == "" {
				continue
			}
			return &ScopeNode{Kind: "func", Name: name}, nil

		case "method_definition":
			methodName := nodeName(n, source)
			parent := findClassParent(n)
			if parent != "" {
				return &ScopeNode{Kind: "method", Name: methodName, Parent: parent}, nil
			}
			return &ScopeNode{Kind: "method", Name: methodName}, nil

		case "class_declaration":
			name := nodeName(n, source)
			if name == "" {
				continue
			}
			return &ScopeNode{Kind: "class", Name: name}, nil

		case "interface_declaration":
			name := nodeName(n, source)
			if name == "" {
				continue
			}
			return &ScopeNode{Kind: "interface", Name: name}, nil

		case "arrow_function":
			parent := findArrowFunctionParent(n)
			if parent != "" {
				return &ScopeNode{Kind: "func", Name: parent}, nil
			}
			return &ScopeNode{Kind: "func", Name: "<anonymous>"}, nil

		case "generator_function_declaration":
			name := nodeName(n, source)
			if name == "" {
				continue
			}
			return &ScopeNode{Kind: "func", Name: name}, nil
		}
	}

	return nil, nil
}

func findClassParent(node *sitter.Node) string {
	for n := node.Parent(); n != nil; n = n.Parent() {
		if n.Kind() == "class_declaration" {
			return n.ChildByFieldName("name").Utf8Text(nil)
		}
	}
	return ""
}

func findArrowFunctionParent(node *sitter.Node) string {
	for n := node.Parent(); n != nil; n = n.Parent() {
		switch n.Kind() {
		case "variable_declarator":
			name := n.ChildByFieldName("name")
			if name != nil {
				return name.Utf8Text(nil)
			}
		case "method_definition":
			name := n.ChildByFieldName("name")
			if name != nil {
				classParent := findClassParent(n)
				if classParent != "" {
					return classParent + "." + name.Utf8Text(nil)
				}
				return name.Utf8Text(nil)
			}
		case "assignment_expression":
			left := n.ChildByFieldName("left")
			if left != nil {
				return left.Utf8Text(nil)
			}
		}
	}
	return ""
}

type tsxAnalyzer struct {
	typescriptAnalyzer
}

func (a *tsxAnalyzer) Language() unsafe.Pointer {
	return ts.LanguageTSX()
}

func init() {
	Register(&typescriptAnalyzer{}, "ts", "js", "jsx")
	Register(&tsxAnalyzer{}, "tsx")
}
