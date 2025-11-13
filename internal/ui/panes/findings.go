package panes

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// FindingsPane displays analysis findings
type FindingsPane struct {
	width    int
	height   int
	findings []*engine.Finding
	selected int
	scroll   int
}

// NewFindingsPane creates a new findings pane
func NewFindingsPane() *FindingsPane {
	return &FindingsPane{
		findings: make([]*engine.Finding, 0),
	}
}

// SetSize sets the pane dimensions
func (p *FindingsPane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetFindings sets the findings data
func (p *FindingsPane) SetFindings(findings []*engine.Finding) {
	p.findings = findings
	if p.selected >= len(findings) {
		p.selected = len(findings) - 1
	}
	if p.selected < 0 {
		p.selected = 0
	}
}

// Update handles messages
func (p *FindingsPane) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if p.selected < len(p.findings)-1 {
				p.selected++
				if p.selected >= p.scroll+p.height {
					p.scroll++
				}
			}
		case "k", "up":
			if p.selected > 0 {
				p.selected--
				if p.selected < p.scroll {
					p.scroll--
				}
			}
		}
	}
	return nil
}

// View renders the pane
func (p *FindingsPane) View() string {
	if len(p.findings) == 0 {
		return "No findings yet"
	}

	var sb strings.Builder

	// Display visible findings
	start := p.scroll
	end := p.scroll + p.height
	if end > len(p.findings) {
		end = len(p.findings)
	}

	for i := start; i < end; i++ {
		finding := p.findings[i]

		// Format finding
		icon := theme.SeverityIcon(string(finding.Severity))
		style := theme.SeverityStyle(string(finding.Severity))

		line := fmt.Sprintf("%s %s:%d - %s",
			icon,
			finding.File,
			finding.LineStart,
			finding.Message,
		)

		// Truncate if too long
		if len(line) > p.width {
			line = line[:p.width-3] + "..."
		}

		if i == p.selected {
			sb.WriteString("▶ " + style.Render(line) + "\n")
		} else {
			sb.WriteString("  " + style.Render(line) + "\n")
		}
	}

	return sb.String()
}

// GetSelected returns the currently selected finding
func (p *FindingsPane) GetSelected() *engine.Finding {
	if p.selected >= 0 && p.selected < len(p.findings) {
		return p.findings[p.selected]
	}
	return nil
}
