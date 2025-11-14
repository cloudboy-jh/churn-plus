package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// PatchPreviewModal shows a patch preview
type PatchPreviewModal struct {
	finding *engine.Finding
	width   int
	height  int
}

// NewPatchPreviewModal creates a new patch preview modal
func NewPatchPreviewModal(finding *engine.Finding) *PatchPreviewModal {
	return &PatchPreviewModal{
		finding: finding,
		width:   80,
		height:  30,
	}
}

// View renders the patch preview modal
func (m *PatchPreviewModal) View() string {
	// Create modal box with solid background
	modalStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.ColorPrimaryRed)).
		Background(lipgloss.Color(theme.ColorBackground)).
		Foreground(lipgloss.Color(theme.ColorTextPrimary)).
		Padding(1, 2).
		Width(m.width).
		Height(m.height)

	var content strings.Builder

	// Title
	title := theme.HighlightStyle.Render("📄 Patch Preview: " + m.finding.File)
	content.WriteString(title)
	content.WriteString("\n\n")

	// Patch content
	patch := m.generatePatch()

	// Style for diff
	diffStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#0d1117")). // Darker background for code
		Foreground(lipgloss.Color(theme.ColorInfo)).
		Padding(1, 2).
		Width(m.width - 8)

	content.WriteString(diffStyle.Render(patch))

	// Footer
	content.WriteString("\n\n")
	footer := theme.MutedStyle.Render("Press 'a' to apply | Press 'q' to close")
	content.WriteString(footer)

	return modalStyle.Render(content.String())
}

// generatePatch generates a unified diff patch
func (m *PatchPreviewModal) generatePatch() string {
	// For now, generate a simple mock patch
	// In a real implementation, this would parse the finding's suggested fix
	// and generate a proper unified diff

	var patch strings.Builder

	patch.WriteString("@@ -")
	patch.WriteString(formatLineRange(m.finding.LineStart, m.finding.LineEnd))
	patch.WriteString(" +")
	patch.WriteString(formatLineRange(m.finding.LineStart, m.finding.LineEnd))
	patch.WriteString(" @@\n")

	if m.finding.Code != "" {
		lines := strings.Split(m.finding.Code, "\n")
		for _, line := range lines {
			// Mark lines to be removed
			if strings.Contains(m.finding.Kind, "unused") ||
				strings.Contains(m.finding.Kind, "unreachable") {
				patch.WriteString(theme.ErrorStyle.Render("-"+line) + "\n")
			} else {
				patch.WriteString(" " + line + "\n")
			}
		}

		// Add suggested fix line
		patch.WriteString(theme.SuccessStyle.Render("+// Fixed by churn-plus") + "\n")
	} else {
		patch.WriteString("  (No code snippet available)\n")
		patch.WriteString(theme.SuccessStyle.Render("+// Fix: ") + m.finding.Message + "\n")
	}

	return patch.String()
}

// formatLineRange formats a line range for diff header
func formatLineRange(start, end int) string {
	if start == end {
		return formatInt(start)
	}
	count := end - start + 1
	return formatInt(start) + "," + formatInt(count)
}

// formatInt formats an integer as a string
func formatInt(n int) string {
	if n < 0 {
		return "0"
	}
	// Use fmt.Sprintf for proper integer to string conversion
	return fmt.Sprintf("%d", n)
}
