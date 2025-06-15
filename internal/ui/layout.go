package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Hint struct {
	Key  string
	Desc string
}

func TermWidth() int {
	const max = 80
	if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
		if w > max {
			return max
		}
		return w
	}
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		if w > max {
			return max
		}
		return w
	}
	return max
}

func RenderFrame(title string, content string, hints []Hint, width int) string {
	var b strings.Builder

	titleText := title + " "
	fill := width - lipgloss.Width(titleText)
	if fill < 1 {
		fill = 1
	}
	b.WriteString(StyleTitle.Render(titleText + strings.Repeat("─", fill)))
	b.WriteString("\n\n")

	b.WriteString(content)
	b.WriteString("\n\n")

	if len(hints) > 0 {
		var parts []string
		for _, h := range hints {
			parts = append(parts, h.Key+" "+h.Desc)
		}
		b.WriteString(StyleHint.Render(strings.Join(parts, " · ")))
	}

	return b.String()
}
