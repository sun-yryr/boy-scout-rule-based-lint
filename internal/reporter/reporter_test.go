package reporter

import (
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
)

type brokenPipeWriter struct{}

func (brokenPipeWriter) Write([]byte) (int, error) {
	return 0, syscall.EPIPE
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

type failAfterFirstWriter struct {
	writes int
}

func (w *failAfterFirstWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == 1 {
		return len(p), nil
	}
	return 0, syscall.EPIPE
}

func TestReporter_Report(t *testing.T) {
	t.Run("counts successful writes", func(t *testing.T) {
		r := NewReporter(io.Discard)
		stop, err := r.Report("main.go:1:1: example")
		if err != nil || stop {
			t.Fatalf("Report() = (%v, %v), want (false, nil)", stop, err)
		}
		if r.NewIssues() != 1 {
			t.Fatalf("NewIssues() = %d, want 1", r.NewIssues())
		}
	})

	t.Run("broken pipe stops without error", func(t *testing.T) {
		r := NewReporter(brokenPipeWriter{})
		stop, err := r.Report("main.go:1:1: example")
		if err != nil {
			t.Fatalf("Report() err = %v, want nil", err)
		}
		if !stop {
			t.Fatal("Report() stop = false, want true")
		}
		if r.NewIssues() != 0 {
			t.Fatalf("NewIssues() = %d, want 0", r.NewIssues())
		}
	})

	t.Run("preserves count on broken pipe after successful writes", func(t *testing.T) {
		r := NewReporter(&failAfterFirstWriter{})

		stop, err := r.Report("first")
		if err != nil || stop {
			t.Fatalf("first Report() = (%v, %v), want (false, nil)", stop, err)
		}

		stop, err = r.Report("second")
		if err != nil {
			t.Fatalf("second Report() err = %v, want nil", err)
		}
		if !stop {
			t.Fatal("second Report() stop = false, want true")
		}
		if r.NewIssues() != 1 {
			t.Fatalf("NewIssues() = %d, want 1", r.NewIssues())
		}
	})

	t.Run("returns genuine write errors", func(t *testing.T) {
		writeErr := errors.New("disk full")
		r := NewReporter(failingWriter{err: writeErr})

		stop, err := r.Report("main.go:1:1: example")
		if stop {
			t.Fatal("Report() stop = true, want false")
		}
		if !errors.Is(err, writeErr) {
			t.Fatalf("Report() err = %v, want %v", err, writeErr)
		}
	})
}

func TestIsBrokenPipe(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "EPIPE", err: syscall.EPIPE, want: true},
		{name: "closed pipe", err: io.ErrClosedPipe, want: true},
		{name: "wrapped EPIPE", err: &os.PathError{Op: "write", Err: syscall.EPIPE}, want: true},
		{name: "other error", err: errors.New("disk full"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBrokenPipe(tt.err); got != tt.want {
				t.Fatalf("isBrokenPipe() = %v, want %v", got, tt.want)
			}
		})
	}
}
