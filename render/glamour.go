package render

import (
	"fmt"
	"regexp"

	"github.com/charmbracelet/glamour"
	glowutils "github.com/charmbracelet/glow/v2/utils"
	"github.com/charmbracelet/lipgloss"
)

// ANSI color codes for temperature bullets.
const (
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiReset  = "\033[0m"
)

// colorTempBullets scans the rendered output for "● XX°C" patterns and
// wraps the bullet with ANSI color codes based on the temperature value.
// This is needed because glamour strips ANSI codes from markdown table cells.
func colorTempBullets(output string) string {
	re := regexp.MustCompile(`(● )(\d{1,3})°C`)
	return re.ReplaceAllStringFunc(output, func(match string) string {
		groups := re.FindStringSubmatch(match)
		if len(groups) < 3 {
			return match
		}
		temp := 0
		fmt.Sscanf(groups[2], "%d", &temp)

		var color string
		switch {
		case temp >= 80:
			color = ansiRed
		case temp >= 70:
			color = ansiYellow
		default:
			color = ansiGreen
		}

		return fmt.Sprintf("%s●%s %s°C", color, ansiReset, groups[2])
	})
}

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

	out, err := r.Render(markdown)
	if err != nil {
		return "", err
	}

	return colorTempBullets(out), nil
}
