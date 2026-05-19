package render

import (
	"github.com/charmbracelet/glamour"
	glowutils "github.com/charmbracelet/glow/v2/utils"
	"github.com/charmbracelet/lipgloss"
)

// Render takes raw markdown, renders it with glamour using the specified style,
// and returns the ANSI-escaped output string.
func Render(markdown, style string, width int) (string, error) {
	opts := []glamour.TermRendererOption{
		glamour.WithColorProfile(lipgloss.ColorProfile()),
		glowutils.GlamourStyle(style, false),
	}

	if width > 0 {
		opts = append(opts, glamour.WithWordWrap(width))
	}

	r, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return "", err
	}

	return r.Render(markdown)
}
