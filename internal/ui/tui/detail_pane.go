package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// DetailPane displays finding details
type DetailPane struct {
	finding *engine.Finding
	width   int
	height  int
}

// NewDetailPane creates a new detail pane
func NewDetailPane() *DetailPane {
	return &DetailPane{}
}

// SetSize sets the pane dimensions
func (p *DetailPane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetFinding sets the finding to display
func (p *DetailPane) SetFinding(finding *engine.Finding) {
	p.finding = finding
}

// View renders the detail pane
func (p *DetailPane) View(focused bool) string {
	// Create border style based on focus
	borderColor := theme.ColorMuted
	if focused {
		borderColor = theme.ColorPrimaryRed
	}

	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		BorderBackground(lipgloss.Color(theme.ColorBackground)).
		Background(lipgloss.Color(theme.ColorBackground)).
		Width(p.width - 2).
		Height(p.height - 2)

	// Create title
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(borderColor)).
		Bold(true).
		Render(" FINDING DETAILS ")

	// Create content
	content := p.renderDetails()

	// Combine title and content
	fullContent := title + "\n" + content

	return borderStyle.Render(fullContent)
}

// renderDetails renders the finding details
func (p *DetailPane) renderDetails() string {
	if p.finding == nil {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.ColorMuted)).
			Background(lipgloss.Color(theme.ColorBackground)).
			Padding(1, 2)

		return emptyStyle.Render("No finding selected")
	}

	var sections []string

	// File info
	sections = append(sections, p.renderFileInfo())
	sections = append(sections, "")

	// Message/Reasoning
	sections = append(sections, p.renderMessage())
	sections = append(sections, "")

	// Code snippet if available
	if p.finding.Code != "" {
		sections = append(sections, p.renderCode())
		sections = append(sections, "")
	}

	// Action buttons
	sections = append(sections, p.renderActions())

	// Wrap content in background style
	contentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.ColorBackground)).
		Foreground(lipgloss.Color(theme.ColorTextPrimary)).
		Padding(0, 2).
		Width(p.width - 8)

	return contentStyle.Render(strings.Join(sections, "\n"))
}

// renderFileInfo renders file and metadata
func (p *DetailPane) renderFileInfo() string {
	var lines []string

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorTextPrimary))

	// File and line
	lines = append(lines, labelStyle.Render("File: ")+valueStyle.Render(
		fmt.Sprintf("%s:%d-%d", p.finding.File, p.finding.LineStart, p.finding.LineEnd),
	))

	// Severity with icon
	icon := theme.SeverityIcon(string(p.finding.Severity))
	severityStyle := theme.SeverityStyle(string(p.finding.Severity))
	lines = append(lines, labelStyle.Render("Severity: ")+severityStyle.Render(
		fmt.Sprintf("%s %s", icon, strings.ToUpper(string(p.finding.Severity))),
	))

	// Type/Kind
	lines = append(lines, labelStyle.Render("Type: ")+valueStyle.Render(p.finding.Kind))

	// Pass
	if p.finding.Pass != "" {
		lines = append(lines, labelStyle.Render("Pass: ")+valueStyle.Render(p.finding.Pass))
	}

	return strings.Join(lines, "\n")
}

// renderMessage renders the finding message
func (p *DetailPane) renderMessage() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorTextPrimary))

	title := labelStyle.Render("Message:")

	// Wrap message to fit width
	wrappedMessage := wrapText(p.finding.Message, p.width-12)
	message := messageStyle.Render(wrappedMessage)

	return title + "\n" + message
}

// renderCode renders the code snippet
func (p *DetailPane) renderCode() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
		Bold(true)

	codeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorInfo)).
		Background(lipgloss.Color("#0d1117")). // Slightly lighter dark for code
		Padding(1, 2).
		Width(p.width - 12)

	title := labelStyle.Render("Code:")
	code := codeStyle.Render(p.finding.Code)

	return title + "\n" + code
}

// renderActions renders the action buttons
func (p *DetailPane) renderActions() string {
	// Create action box
	actionStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.ColorSecondaryRed)).
		Background(lipgloss.Color(theme.ColorBackground)).
		Foreground(lipgloss.Color(theme.ColorTextPrimary)).
		Padding(0, 2).
		Width(p.width - 12)

	actions := []string{
		fmt.Sprintf("%s Send to LLM  %s",
			theme.HighlightStyle.Render("(l)"),
			theme.MutedStyle.Render("← PRIMARY ACTION")),
		fmt.Sprintf("%s Preview Patch", theme.HighlightStyle.Render("(p)")),
		fmt.Sprintf("%s Apply Patch", theme.HighlightStyle.Render("(a)")),
		fmt.Sprintf("%s Back to Menu", theme.MutedStyle.Render("(m)")),
		fmt.Sprintf("%s Back to List", theme.MutedStyle.Render("(q)")),
	}

	return actionStyle.Render(strings.Join(actions, "\n"))
}

// wrapText wraps text to fit within a specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Word itself is longer than width
				lines = append(lines, word)
				currentLine = ""
			}
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}
