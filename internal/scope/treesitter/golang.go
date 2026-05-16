package treesitter

import (
	"unsafe"

	sitter "github.com/tree-sitter/go-tree-sitter"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

type goAnalyzer struct{}

func (a *goAnalyzer) Language() unsafe.Pointer {
	return golang.Language()
}

func (a *goAnalyzer) FindScope(node *sitter.Node, source []byte) (*ScopeNode, error) {
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

		case "method_declaration":
			methodName := nodeName(n, source)
			receiver := n.ChildByFieldName("receiver")
			receiverName := extractReceiverType(receiver, source)
			if receiverName != "" {
				return &ScopeNode{Kind: "method", Name: methodName, Parent: receiverName}, nil
			}
			return &ScopeNode{Kind: "method", Name: methodName}, nil

		case "type_spec":
			switch e := n.Parent(); {
			case e != nil && e.Kind() == "type_declaration":
				name := nodeName(n, source)
				if name == "" {
					continue
				}
				typeNode := n.ChildByFieldName("type")
				if typeNode == nil {
					continue
				}
				typeKind := extractTypeKind(typeNode)
				return &ScopeNode{Kind: typeKind, Name: name}, nil
			}
		}
	}

	return nil, nil
}

func extractReceiverType(receiver *sitter.Node, source []byte) string {
	if receiver == nil {
		return ""
	}
	for i := uint(0); i < receiver.NamedChildCount(); i++ {
		child := receiver.NamedChild(i)
		if child.Kind() == "parameter_declaration" {
			typeChild := child.ChildByFieldName("type")
			if typeChild == nil {
				continue
			}
			name := unwrapPointerType(typeChild, source)
			if name != "" {
				return name
			}
		}
	}
	return ""
}

func unwrapPointerType(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	if node.Kind() == "pointer_type" {
		for i := uint(0); i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if child.Kind() == "type_identifier" {
				return child.Utf8Text(source)
			}
		}
	}
	if node.Kind() == "type_identifier" {
		return node.Utf8Text(source)
	}
	return node.Utf8Text(source)
}

func extractTypeKind(typeNode *sitter.Node) string {
	if typeNode == nil {
		return "type"
	}
	kind := typeNode.Kind()
	switch kind {
	case "struct_type":
		return "struct"
	case "interface_type":
		return "interface"
	default:
		return "type"
	}
}

func init() {
	Register(&goAnalyzer{}, "go")
}
