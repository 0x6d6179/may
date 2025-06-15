package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func mayTheme() *huh.Theme {
	t := huh.ThemeCharm()

	lavender := lipgloss.Color("#C9B8FF")
	rose := lipgloss.Color("#FFB3C6")
	mint := lipgloss.Color("#A8F5CF")
	sky := lipgloss.Color("#B8D9FF")

	t.Focused.Title = t.Focused.Title.Foreground(lavender)
	t.Blurred.Title = t.Blurred.Title.Foreground(lavender)

	t.Focused.SelectSelector = t.Focused.SelectSelector.SetString("→ ").Foreground(rose)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.SetString("→ ").Foreground(rose)

	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(mint)
	t.Focused.Description = t.Focused.Description.Foreground(sky)

	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.SetString("◆ ")
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.SetString("◇ ")
	t.Blurred.SelectedPrefix = t.Blurred.SelectedPrefix.SetString("◆ ")
	t.Blurred.UnselectedPrefix = t.Blurred.UnselectedPrefix.SetString("◇ ")

	return t
}

// NewForm returns a huh.Form with project-wide defaults applied.
// All forms in may must use this instead of huh.NewForm directly.
func NewForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).WithTheme(mayTheme())
}

// NewSelect returns a huh.Select with project-wide defaults (filtering on).
// All selects in may must use this instead of huh.NewSelect directly.
func NewSelect[T comparable]() *huh.Select[T] {
	return huh.NewSelect[T]().Filtering(true)
}

func NewConfirm() *huh.Confirm {
	return huh.NewConfirm().WithButtonAlignment(lipgloss.Left)
}

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

// Header prints a lavender-colored "title ─────────────────" line to w,
// filling to terminal width. Call immediately before form.Run().
func Header(w io.Writer, title string) {
	width := termWidth()
	fill := width - len(title) - 1
	if fill < 1 {
		fill = 1
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#C9B8FF"))
	fmt.Fprintln(w, style.Render(title+" "+strings.Repeat("─", fill)))
}

func termWidth() int {
	const max = 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		if w > max {
			return max
		}
		return w
	}
	if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
		if w > max {
			return max
		}
		return w
	}
	return max
}
