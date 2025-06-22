package ui

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Spinner starts a braille spinner on w with the given label.
// Call the returned stop func when the operation completes; it clears the line.
// stop() blocks until the line is cleared — no race with subsequent renders.
func Spinner(w io.Writer, label string) (stop func()) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	quit := make(chan struct{})
	cleared := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-quit:
				fmt.Fprint(w, "\r\033[K")
				close(cleared)
				return
			default:
				fmt.Fprintf(w, "\r%s %s", frames[i%len(frames)], label)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	return func() {
		close(quit)
		<-cleared
	}
}

// NewTable returns a tabwriter.Writer with project-wide consistent settings.
func NewTable(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}
