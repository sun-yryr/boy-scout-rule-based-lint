package pathnorm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	repoRootOnce sync.Once
	repoRoot     string
	repoRootErr  error
)

func gitRepoRoot() (string, error) {
	repoRootOnce.Do(func() {
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		out, err := cmd.Output()
		if err != nil {
			repoRootErr = err
			return
		}
		repoRoot = strings.TrimSpace(string(out))
	})
	return repoRoot, repoRootErr
}

// Normalize returns a repo-root-relative path using forward slashes.
// Paths outside the repository are rejected.
// When git is unavailable, it falls back to a cleaned slash-separated path
// that must not traverse above the working directory.
func Normalize(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("empty file path")
	}

	root, err := gitRepoRoot()
	if err != nil {
		cleaned := cleanSlashPath(filePath)
		if err := rejectTraversal(cleaned); err != nil {
			return "", err
		}
		return cleaned, nil
	}

	abs, err := toAbsolute(filePath, root)
	if err != nil {
		return "", err
	}

	rel, err := repoRelativePath(root, abs)
	if err != nil {
		return "", fmt.Errorf("normalizing %q: %w", filePath, err)
	}
	return rel, nil
}

// Resolve converts a repo-relative path to an absolute path for file I/O.
// The resolved path must stay within the repository root.
func Resolve(repoRelative string) (string, error) {
	if repoRelative == "" {
		return "", fmt.Errorf("empty file path")
	}
	if filepath.IsAbs(repoRelative) {
		return "", fmt.Errorf("absolute paths are not allowed: %q", repoRelative)
	}
	if err := rejectTraversal(repoRelative); err != nil {
		return "", err
	}

	root, err := gitRepoRoot()
	if err != nil {
		return cleanSlashPath(repoRelative), nil
	}

	abs := filepath.Join(root, filepath.FromSlash(repoRelative))
	abs = filepath.Clean(abs)
	if !isWithinRoot(root, abs) {
		return "", fmt.Errorf("path %q is outside repository root", repoRelative)
	}
	return abs, nil
}

func toAbsolute(filePath, repoRoot string) (string, error) {
	if filepath.IsAbs(filePath) {
		return filepath.Clean(filePath), nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(cwd, filePath)), nil
}

func cleanSlashPath(filePath string) string {
	p := filepath.ToSlash(filepath.Clean(filePath))
	return strings.TrimPrefix(p, "./")
}

func repoRelativePath(root, abs string) (string, error) {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if !isWithinRoot(root, abs) {
		return "", fmt.Errorf("path is outside repository root")
	}
	return rel, nil
}

func isWithinRoot(root, target string) bool {
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func rejectTraversal(path string) error {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("path %q escapes allowed directory", path)
	}
	return nil
}
