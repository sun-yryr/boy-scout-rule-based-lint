package diff

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type ChangeSet struct {
	Files        map[string]bool
	ChangedLines map[string]map[int]bool
}

func (cs *ChangeSet) HasFile(path string) bool {
	return cs.Files[path]
}

func (cs *ChangeSet) HasLine(path string, line int) bool {
	if !cs.Files[path] {
		return false
	}
	lines := cs.ChangedLines[path]
	if len(lines) == 0 {
		return false
	}
	return lines[line]
}

func ParseDiff(r io.Reader) (*ChangeSet, error) {
	cs := &ChangeSet{
		Files:        make(map[string]bool),
		ChangedLines: make(map[string]map[int]bool),
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	var currentFile string
	var newLine int
	var inHunk bool

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "+++ ") {
			currentFile = extractNewFilePath(line)
			if currentFile != "" {
				cs.Files[currentFile] = true
				if cs.ChangedLines[currentFile] == nil {
					cs.ChangedLines[currentFile] = make(map[int]bool)
				}
			}
			inHunk = false
			continue
		}

		if isDiffHeader(line) {
			inHunk = false
			continue
		}

		if strings.HasPrefix(line, "@@") {
			newLine = parseHunkNewStart(line)
			inHunk = true
			continue
		}

		if !inHunk || currentFile == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "+"):
			cs.ChangedLines[currentFile][newLine] = true
			newLine++
		case strings.HasPrefix(line, " "):
			newLine++
		}
	}

	return cs, scanner.Err()
}

func extractNewFilePath(line string) string {
	rest := strings.TrimPrefix(line, "+++ ")
	if rest == "/dev/null" {
		return ""
	}
	rest = stripTrailer(rest)
	if path, ok := strings.CutPrefix(rest, "b/"); ok {
		return path
	}
	if path, ok := strings.CutPrefix(rest, "a/"); ok {
		return path
	}
	return rest
}

func stripTrailer(s string) string {
	if idx := strings.IndexByte(s, '\t'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func parseHunkNewStart(line string) int {
	plusIdx := strings.Index(line, "+")
	if plusIdx < 0 {
		return 0
	}
	rest := line[plusIdx+1:]
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0
	}
	n, err := strconv.Atoi(rest[:i])
	if err != nil {
		return 0
	}
	return n
}

func isDiffHeader(line string) bool {
	return strings.HasPrefix(line, "diff ") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "--- ") ||
		strings.HasPrefix(line, "new file ") ||
		strings.HasPrefix(line, "deleted file ") ||
		strings.HasPrefix(line, "rename ") ||
		strings.HasPrefix(line, "copy ") ||
		strings.HasPrefix(line, "similarity ") ||
		strings.HasPrefix(line, "old mode ") ||
		strings.HasPrefix(line, "new mode ") ||
		strings.HasPrefix(line, "Binary files ")
}

func GetDiff(baseRef string) (*ChangeSet, error) {
	cmd := exec.Command("git", "diff", "--unified=0", baseRef+"...HEAD")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("running git diff: %w", err)
	}
	cs, err := ParseDiff(stdout)
	if err != nil {
		if waitErr := cmd.Wait(); waitErr != nil {
			return nil, fmt.Errorf("parsing diff: %w; git wait: %w", err, waitErr)
		}
		return nil, fmt.Errorf("parsing diff: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	return cs, nil
}
