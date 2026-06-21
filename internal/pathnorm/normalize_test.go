package pathnorm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalize_FallbackWithoutGit(t *testing.T) {
	// Temp dirs are typically outside a git work tree; Normalize falls back to cleaning.
	tmpDir := t.TempDir()
	chdirTo(t, tmpDir)

	tests := []struct {
		in   string
		want string
	}{
		{"main.go", "main.go"},
		{"./main.go", "main.go"},
		{"internal/foo.go", "internal/foo.go"},
	}

	for _, tt := range tests {
		got, err := Normalize(tt.in)
		if err != nil {
			t.Fatalf("Normalize(%q) err = %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalize_InGitRepo(t *testing.T) {
	repoRoot, err := gitRepoRoot()
	if err != nil {
		t.Skip("not in a git repository")
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	absMain := filepath.Join(repoRoot, "cmd", "bsr", "main.go")
	got, err := Normalize(absMain)
	if err != nil {
		t.Fatalf("Normalize(abs) err = %v", err)
	}
	want := "cmd/bsr/main.go"
	if got != want {
		t.Errorf("Normalize(abs) = %q, want %q", got, want)
	}

	got, err = Normalize("./cmd/bsr/main.go")
	if err != nil {
		t.Fatalf("Normalize(rel) err = %v", err)
	}
	if got != want {
		t.Errorf("Normalize(rel) = %q, want %q", got, want)
	}
}

func TestNormalize_RejectsOutsideRepo(t *testing.T) {
	repoRoot, err := gitRepoRoot()
	if err != nil {
		t.Skip("not in a git repository")
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	outside := filepath.Join(repoRoot, "..", "outside-bsr-test.go")
	if _, err := Normalize(outside); err == nil {
		t.Fatal("Normalize(outside repo) err = nil, want error")
	}
	if _, err := Normalize("/etc/passwd"); err == nil {
		t.Fatal("Normalize(absolute outside repo) err = nil, want error")
	}
}

func TestNormalize_RejectsTraversalWithoutGit(t *testing.T) {
	tmpDir := t.TempDir()
	chdirTo(t, tmpDir)

	if _, err := Normalize("../escape.go"); err == nil {
		t.Fatal("Normalize(../escape.go) err = nil, want error")
	}
}

func TestResolve_RejectsTraversal(t *testing.T) {
	repoRoot, err := gitRepoRoot()
	if err != nil {
		t.Skip("not in a git repository")
	}

	if _, err := Resolve("../escape.go"); err == nil {
		t.Fatal("Resolve(../escape.go) err = nil, want error")
	}
	if _, err := Resolve("/etc/passwd"); err == nil {
		t.Fatal("Resolve(absolute path) err = nil, want error")
	}

	got, err := Resolve("cmd/bsr/main.go")
	if err != nil {
		t.Fatalf("Resolve(in-repo) err = %v", err)
	}
	want := filepath.Join(repoRoot, "cmd", "bsr", "main.go")
	if got != want {
		t.Errorf("Resolve(in-repo) = %q, want %q", got, want)
	}
}

func TestResolve_FallbackWithoutGit(t *testing.T) {
	got, err := Resolve("main.go")
	if err != nil {
		t.Fatalf("Resolve() err = %v", err)
	}
	if got != "main.go" {
		t.Errorf("Resolve() = %q, want main.go", got)
	}
}

func chdirTo(t *testing.T, dir string) {
	t.Helper()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q): %v", dir, err)
	}
}
