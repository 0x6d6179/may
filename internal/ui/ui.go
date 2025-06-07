package ui

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/huh"
)

// NewForm returns a huh.Form with project-wide defaults applied.
// All forms in may must use this instead of huh.NewForm directly.
func NewForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).WithHeight(10)
}

// NewSelect returns a huh.Select with project-wide defaults (filtering on).
// All selects in may must use this instead of huh.NewSelect directly.
func NewSelect[T comparable]() *huh.Select[T] {
	return huh.NewSelect[T]().Filtering(true)
}

// Spinner starts a braille spinner on w with the given label.
// Call the returned stop func when the operation completes; it clears the line.
func Spinner(w io.Writer, label string) (stop func()) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-done:
				fmt.Fprint(w, "\r                    \r")
				return
			default:
				fmt.Fprintf(w, "\r%s %s", frames[i%len(frames)], label)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	return func() {
		close(done)
		time.Sleep(10 * time.Millisecond)
	}
}

// NewTable returns a tabwriter.Writer with project-wide consistent settings.
func NewTable(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}
