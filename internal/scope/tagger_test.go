package scope

import (
	"testing"
)

func TestLanguageFromExt(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"foo.go", "go"},
		{"foo.py", "python"},
		{"foo.js", "javascript"},
		{"foo.ts", "typescript"},
		{"foo.tsx", "typescript"},
		{"foo.php", "php"},
		{"path/to/file.rs", "rust"},
		{"file.unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := languageFromExt(tt.path)
			if got != tt.want {
				t.Errorf("languageFromExt(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsScopeKind(t *testing.T) {
	if !isScopeKind("func") {
		t.Error("func should be a scope kind")
	}
	if !isScopeKind("class") {
		t.Error("class should be a scope kind")
	}
	if !isScopeKind("struct") {
		t.Error("struct should be a scope kind")
	}
	if isScopeKind("variable") {
		t.Error("variable should not be a scope kind")
	}
	if isScopeKind("field") {
		t.Error("field should not be a scope kind")
	}
}

func TestFindEnclosing(t *testing.T) {
	tags := []tag{
		{Name: "foo", Kind: "func", Line: 5, End: 10},
		{Name: "Bar", Kind: "struct", Line: 12, End: 30},
		{Name: "method", Kind: "func", Line: 15, End: 25, Scope: "Bar"},
	}

	tests := []struct {
		line     int
		wantName string
		wantNil  bool
	}{
		{line: 3, wantNil: true},
		{line: 5, wantName: "foo"},
		{line: 8, wantName: "foo"},
		{line: 15, wantName: "method"},
		{line: 20, wantName: "method"},
		{line: 27, wantName: "Bar"},
		{line: 13, wantName: "Bar"},
		{line: 35, wantNil: true},
	}

	for _, tt := range tests {
		got := findEnclosing(tags, tt.line)
		if tt.wantNil {
			if got != nil {
				t.Errorf("findEnclosing(line=%d) = %v, want nil", tt.line, got)
			}
		} else {
			if got == nil {
				t.Errorf("findEnclosing(line=%d) = nil, want %q", tt.line, tt.wantName)
			} else if got.Name != tt.wantName {
				t.Errorf("findEnclosing(line=%d).Name = %q, want %q", tt.line, got.Name, tt.wantName)
			}
		}
	}
}

func TestFormatScope(t *testing.T) {
	tests := []struct {
		filePath string
		tag      tag
		want     string
	}{
		{
			filePath: "sample.go",
			tag:      tag{Name: "foo", Kind: "func"},
			want:     "go:func:foo",
		},
		{
			filePath: "sample.go",
			tag:      tag{Name: "Extract", Kind: "func", Scope: "Extractor"},
			want:     "go:func:Extractor.Extract",
		},
		{
			filePath: "sample.py",
			tag:      tag{Name: "MyClass", Kind: "class"},
			want:     "python:class:MyClass",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatScope(tt.filePath, &tt.tag)
			if got != tt.want {
				t.Errorf("formatScope() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFileScope(t *testing.T) {
	got := fileScope("foo.go")
	if got != "go:file" {
		t.Errorf("fileScope(foo.go) = %q, want %q", got, "go:file")
	}
}
