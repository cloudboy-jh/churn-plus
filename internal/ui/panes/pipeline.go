package panes

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// PipelinePane displays pipeline execution status
type PipelinePane struct {
	width    int
	height   int
	pipeline *engine.Pipeline
	scroll   int
}

// NewPipelinePane creates a new pipeline pane
func NewPipelinePane() *PipelinePane {
	return &PipelinePane{}
}

// SetSize sets the pane dimensions
func (p *PipelinePane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// UpdatePipeline updates the pipeline data
func (p *PipelinePane) UpdatePipeline(pipeline *engine.Pipeline) {
	p.pipeline = pipeline
}

// Update handles messages
func (p *PipelinePane) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if p.scroll < len(p.pipeline.Passes)-p.height {
				p.scroll++
			}
		case "k", "up":
			if p.scroll > 0 {
				p.scroll--
			}
		}
	}
	return nil
}

// View renders the pane
func (p *PipelinePane) View() string {
	if p.pipeline == nil || len(p.pipeline.Passes) == 0 {
		return "Pipeline not started"
	}

	var sb strings.Builder

	// Display passes
	for i, pass := range p.pipeline.Passes {
		if i < p.scroll {
			continue
		}
		if i >= p.scroll+p.height {
			break
		}

		icon := theme.StatusIcon(string(pass.Status))
		style := theme.MutedStyle

		switch pass.Status {
		case engine.PassCompleted:
			style = theme.SuccessStyle
		case engine.PassFailed:
			style = theme.ErrorStyle
		case engine.PassRunning:
			style = theme.InfoStyle
		}

		line := fmt.Sprintf("%s %s - %s (%s)",
			icon,
			pass.Name,
			pass.Description,
			pass.Model,
		)

		sb.WriteString(style.Render(line) + "\n")

		// Show error if failed
		if pass.Status == engine.PassFailed && pass.Error != "" {
			errorLine := fmt.Sprintf("   Error: %s", pass.Error)
			if len(errorLine) > p.width {
				errorLine = errorLine[:p.width-3] + "..."
			}
			sb.WriteString(theme.ErrorStyle.Render(errorLine) + "\n")
		}
	}

	// Show summary
	if p.pipeline.EndTime.IsZero() {
		sb.WriteString("\n" + theme.InfoStyle.Render("Pipeline running...") + "\n")
	} else {
		duration := p.pipeline.EndTime.Sub(p.pipeline.StartTime)
		summary := fmt.Sprintf("\nCompleted in %s", duration.Round(100))
		sb.WriteString(theme.SuccessStyle.Render(summary) + "\n")
	}

	return sb.String()
}
