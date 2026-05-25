package reporter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
)

// Reporter writes new issue lines to stdout and tracks how many were emitted.
type Reporter struct {
	stdout    io.Writer
	newIssues int
}

// NewReporter creates a Reporter that writes to stdout.
func NewReporter(stdout io.Writer) *Reporter {
	return &Reporter{stdout: stdout}
}

// NewIssues returns the number of issue lines successfully written.
func (r *Reporter) NewIssues() int {
	return r.newIssues
}

// Report emits one new issue line.
// stop=true means the downstream reader closed stdout (broken pipe); this is not an error.
func (r *Reporter) Report(line string) (stop bool, err error) {
	if _, err := fmt.Fprintln(r.stdout, line); err != nil {
		if isBrokenPipe(err) {
			return true, nil
		}
		return false, err
	}
	r.newIssues++
	return false, nil
}

func isBrokenPipe(err error) bool {
	if errors.Is(err, io.ErrClosedPipe) {
		return true
	}
	if errors.Is(err, syscall.EPIPE) {
		return true
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return isBrokenPipe(pathErr.Err)
	}
	return false
}
