package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette from original Churn
var (
	// Core colors
	ColorBackground   = lipgloss.Color("#1b1b1b")
	ColorPrimaryRed   = lipgloss.Color("#ff5656")
	ColorSecondaryRed = lipgloss.Color("#ff8585")
	ColorTextPrimary  = lipgloss.Color("#f2e9e4")
	ColorMuted        = lipgloss.Color("#a6adc8")

	// Status colors
	ColorInfo    = lipgloss.Color("#8ab4f8")
	ColorSuccess = lipgloss.Color("#a6e3a1")
	ColorWarning = lipgloss.Color("#f9e2af")
	ColorError   = lipgloss.Color("#f38ba8")
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorBackground)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimaryRed).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorPrimaryRed).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)
)

// Pane styles
var (
	PaneBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)

	ActivePaneBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimaryRed).
				Padding(0, 1)

	PaneTitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimaryRed).
			Bold(true).
			Padding(0, 1)
)

// ASCII Logo - CHURN in retro pixel style matching original design
const logoRaw = `
 ██████   ██  ██   ██  ██   ██  ██████   ██████
██   ██   ██  ██   ██  ██   ██  ██   ██  ██   ██
██        ██████   ██  ██   ██  ██████   ██   ██
██   ██   ██  ██   ██  ██   ██  ██  ██   ██   ██
 ██████   ██  ██    ████     ██  ██       ██████`

// RedGradient applies a red gradient effect to the logo
// Ported from original Churn's redGradient function
func RedGradient(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	// Create gradient from primary to secondary red
	totalLines := len(lines)

	for i, line := range lines {
		// Calculate gradient position (0.0 to 1.0)
		position := float64(i) / float64(totalLines)

		// Interpolate between primary and secondary red
		var color lipgloss.Color
		if position < 0.5 {
			color = ColorPrimaryRed
		} else {
			color = ColorSecondaryRed
		}

		style := lipgloss.NewStyle().Foreground(color)
		result.WriteString(style.Render(line))
		if i < totalLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// RenderLogo returns the styled ASCII logo
func RenderLogo() string {
	return RedGradient(logoRaw)
}

// Box drawing characters (matching original Churn)
const (
	BoxVertical       = "│"
	BoxHorizontal     = "─"
	BoxTopLeft        = "┌"
	BoxTopRight       = "┐"
	BoxBottomLeft     = "└"
	BoxBottomRight    = "┘"
	BoxVerticalRight  = "├"
	BoxVerticalLeft   = "┤"
	BoxHorizontalDown = "┬"
	BoxHorizontalUp   = "┴"
	BoxCross          = "┼"
)

// Progress bar characters (matching original Churn)
const (
	ProgressFull    = "█"
	ProgressEmpty   = "░"
	ProgressPartial = "▓"
)

// CreateProgressBar creates a terminal progress bar
func CreateProgressBar(current, total, width int) string {
	if total == 0 {
		return strings.Repeat(ProgressEmpty, width)
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat(ProgressFull, filled)
	empty := strings.Repeat(ProgressEmpty, width-filled)

	return SuccessStyle.Render(bar) + MutedStyle.Render(empty)
}

// SeverityStyle returns the appropriate style for a severity level
func SeverityStyle(severity string) lipgloss.Style {
	switch severity {
	case "critical", "high":
		return ErrorStyle
	case "medium":
		return WarningStyle
	case "low":
		return InfoStyle
	default:
		return MutedStyle
	}
}

// SeverityIcon returns an icon for a severity level
func SeverityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🔵"
	default:
		return "⚪"
	}
}

// StatusIcon returns an icon for pass status
func StatusIcon(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "running":
		return "⚙️"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "⚪"
	}
}
