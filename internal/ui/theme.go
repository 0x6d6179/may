package ui

import "github.com/charmbracelet/lipgloss"

var (
	lavender = lipgloss.Color("#C9B8FF")
	rose     = lipgloss.Color("#FFB3C6")
	mint     = lipgloss.Color("#A8F5CF")
	sky      = lipgloss.Color("#B8D9FF")

	StyleTitle       = lipgloss.NewStyle().Foreground(lavender).Bold(true)
	StyleSelector    = lipgloss.NewStyle().Foreground(rose)
	StyleSelected    = lipgloss.NewStyle().Foreground(mint)
	StyleDescription = lipgloss.NewStyle().Foreground(sky)
	StyleSeparator   = lipgloss.NewStyle().Foreground(lavender)
	StyleHint        = lipgloss.NewStyle().Faint(true)
	StyleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
)
