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
// When git is unavailable, it falls back to a cleaned slash-separated path.
func Normalize(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("empty file path")
	}

	root, err := gitRepoRoot()
	if err != nil {
		return cleanSlashPath(filePath), nil
	}

	abs, err := toAbsolute(filePath, root)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}

	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "../") {
		return cleanSlashPath(filePath), nil
	}
	return rel, nil
}

// Resolve converts a repo-relative path to an absolute path for file I/O.
func Resolve(repoRelative string) (string, error) {
	root, err := gitRepoRoot()
	if err != nil {
		return repoRelative, nil
	}
	return filepath.Join(root, filepath.FromSlash(repoRelative)), nil
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
