package iostreams

import (
	"io"
	"os"

	"golang.org/x/term"
)

// IOStreams holds the three standard streams for a command invocation.
// Out is the stdout contract for ws/wt cd-path output — nothing else goes here.
// ErrOut is for all human-readable text: prompts, errors, status messages.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// System returns IOStreams wired to the process's standard file descriptors.
func System() *IOStreams {
	return &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

// Test returns IOStreams wired to the given writers, for use in tests.
func Test(in io.Reader, out, errOut io.Writer) *IOStreams {
	return &IOStreams{In: in, Out: out, ErrOut: errOut}
}

// IsTerminal reports whether Out is connected to an interactive terminal.
func (s *IOStreams) IsTerminal() bool {
	f, ok := s.Out.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// IsErrTerminal reports whether ErrOut is connected to an interactive terminal.
// Use this instead of IsTerminal when stdout may be captured (e.g. shell wrapper functions).
func (s *IOStreams) IsErrTerminal() bool {
	f, ok := s.ErrOut.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// ColorEnabled reports whether terminal color output is appropriate.
func (s *IOStreams) ColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return s.IsTerminal()
}
