package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// ListPane displays the findings list
type ListPane struct {
	findings []*engine.Finding
	selected int
	scroll   int
	width    int
	height   int
}

// NewListPane creates a new list pane
func NewListPane(findings []*engine.Finding) *ListPane {
	return &ListPane{
		findings: findings,
		selected: 0,
		scroll:   0,
	}
}

// SetSize sets the pane dimensions
func (p *ListPane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetSelected sets the selected index
func (p *ListPane) SetSelected(idx int) {
	p.selected = idx

	// Adjust scroll if needed
	visibleCount := p.height - 4 // Account for title and borders
	if p.selected < p.scroll {
		p.scroll = p.selected
	} else if p.selected >= p.scroll+visibleCount {
		p.scroll = p.selected - visibleCount + 1
	}
}

// View renders the list pane
func (p *ListPane) View(focused bool) string {
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
		Render(fmt.Sprintf(" FINDINGS (%d) ", len(p.findings)))

	// Create content
	content := p.renderFindings()

	// Combine title and content
	fullContent := title + "\n" + content

	return borderStyle.Render(fullContent)
}

// renderFindings renders the findings list
func (p *ListPane) renderFindings() string {
	if len(p.findings) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.ColorMuted)).
			Background(lipgloss.Color(theme.ColorBackground)).
			Padding(1, 2)

		return emptyStyle.Render("No findings to display\n\nRun a scan first")
	}

	var items []string

	// Calculate visible range
	visibleCount := p.height - 4
	start := p.scroll
	end := start + visibleCount
	if end > len(p.findings) {
		end = len(p.findings)
	}

	// Render visible findings
	for i := start; i < end; i++ {
		finding := p.findings[i]
		items = append(items, p.renderFindingItem(finding, i == p.selected))
	}

	return strings.Join(items, "\n")
}

// renderFindingItem renders a single finding item
func (p *ListPane) renderFindingItem(finding *engine.Finding, isSelected bool) string {
	// Get severity icon
	icon := theme.SeverityIcon(string(finding.Severity))

	// Create short label
	fileName := finding.File
	if len(fileName) > 20 {
		// Truncate long filenames
		fileName = "..." + fileName[len(fileName)-17:]
	}

	label := fmt.Sprintf("%s %s:%d", icon, fileName, finding.LineStart)

	// Truncate if too long
	maxWidth := p.width - 8
	if len(label) > maxWidth {
		label = label[:maxWidth-3] + "..."
	}

	if isSelected {
		// Selected item with solid coral background
		selectedStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.ColorPrimaryRed)).
			Foreground(lipgloss.Color(theme.ColorTextPrimary)).
			Bold(true).
			Padding(0, 1).
			Width(p.width - 6)

		return selectedStyle.Render("▶ " + label)
	} else {
		// Unselected item with dark background
		unselectedStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.ColorBackground)).
			Foreground(lipgloss.Color(theme.ColorTextPrimary)).
			Padding(0, 1).
			Width(p.width - 6)

		return unselectedStyle.Render("  " + label)
	}
}
