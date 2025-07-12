package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

const AILogo = "⚹"

var (
	lavender = lipgloss.Color("#C9B8FF")
	rose     = lipgloss.Color("#FFB3C6")
	mint     = lipgloss.Color("#A8F5CF")
	sky      = lipgloss.Color("#B8D9FF")

	safetyGreen  = lipgloss.Color("#4CD964")
	safetyYellow = lipgloss.Color("#FFD93D")
	safetyOrange = lipgloss.Color("#FFB86B")
	safetyRed    = lipgloss.Color("#FF6B6B")

	StyleTitle       = lipgloss.NewStyle().Foreground(lavender).Bold(true)
	StyleSelector    = lipgloss.NewStyle().Foreground(rose)
	StyleSelected    = lipgloss.NewStyle().Foreground(mint)
	StyleDescription = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	StyleSeparator   = lipgloss.NewStyle().Foreground(lavender)
	StyleHint        = lipgloss.NewStyle().Faint(true)
	StyleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	StyleSafetySafe        = lipgloss.NewStyle().Foreground(safetyGreen)
	StyleSafetyCautious    = lipgloss.NewStyle().Foreground(safetyYellow)
	StyleSafetyRisky       = lipgloss.NewStyle().Foreground(safetyOrange)
	StyleSafetyDestructive = lipgloss.NewStyle().Foreground(safetyRed)
	StyleWarning           = lipgloss.NewStyle().Foreground(safetyRed).Bold(true)

	AISpinner = spinner.Spinner{
		Frames: []string{
			lipgloss.NewStyle().Foreground(lavender).Render(AILogo),
			lipgloss.NewStyle().Foreground(rose).Render(AILogo),
			lipgloss.NewStyle().Foreground(mint).Render(AILogo),
			lipgloss.NewStyle().Foreground(sky).Render(AILogo),
		},
		FPS: 200 * time.Millisecond,
	}
)
