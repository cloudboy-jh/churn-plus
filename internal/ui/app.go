package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
	"github.com/cloudboy-jh/churn-plus/internal/ui/panes"
)

// FocusPane represents which pane has focus
type FocusPane int

const (
	FocusPaneFileTree FocusPane = iota
	FocusPaneCodeView
	FocusPaneFindings
	FocusPanePipeline
)

// Model is the main BubbleTea model
type Model struct {
	width  int
	height int

	// Panes
	fileTree *panes.FileTreePane
	codeView *panes.CodeViewPane
	findings *panes.FindingsPane
	pipeline *panes.PipelinePane

	// State
	focus        FocusPane
	fileTreeData *engine.FileNode
	findingsData []*engine.Finding
	pipelineData *engine.Pipeline

	// Pipeline events channel
	events <-chan engine.PipelineEvent

	ready bool
}

// NewModel creates a new TUI model
func NewModel(fileTree *engine.FileNode, pipeline *engine.Pipeline, orchestrator *engine.PipelineOrchestrator, files []*engine.FileInfo, ctx *engine.ProjectContext, autoRun bool) Model {
	return Model{
		fileTree:     panes.NewFileTreePane(),
		codeView:     panes.NewCodeViewPane(),
		findings:     panes.NewFindingsPane(),
		pipeline:     panes.NewPipelinePane(),
		focus:        FocusPaneFileTree,
		fileTreeData: fileTree,
		findingsData: make([]*engine.Finding, 0),
		pipelineData: pipeline,
		events:       nil, // Will be set when pipeline starts
		ready:        false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update pane sizes
		m.updatePaneSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Cycle focus
			m.focus = (m.focus + 1) % 4

		case "h", "left":
			if m.focus == FocusPaneCodeView {
				m.focus = FocusPaneFileTree
			} else if m.focus == FocusPanePipeline {
				m.focus = FocusPaneFindings
			}

		case "l", "right":
			if m.focus == FocusPaneFileTree {
				m.focus = FocusPaneCodeView
			} else if m.focus == FocusPaneFindings {
				m.focus = FocusPanePipeline
			}

		case "j", "down":
			if m.focus == FocusPaneFileTree {
				m.focus = FocusPaneFindings
			} else if m.focus == FocusPaneCodeView {
				m.focus = FocusPanePipeline
			}

		case "k", "up":
			if m.focus == FocusPaneFindings {
				m.focus = FocusPaneFileTree
			} else if m.focus == FocusPanePipeline {
				m.focus = FocusPaneCodeView
			}

		case " ":
			// Space to trigger actions in focused pane
			// For now, just a placeholder
		}

	case PipelineStartedMsg:
		// Pipeline has started, set up event listening
		m.events = msg.events
		cmds = append(cmds, waitForPipelineEvent(m.events))

	case PipelineEventMsg:
		// Handle pipeline events
		switch msg.Event.Type {
		case engine.EventFindingAdded:
			m.findingsData = append(m.findingsData, msg.Event.Finding)
			m.findings.SetFindings(m.findingsData)
		case engine.EventPassStarted, engine.EventPassCompleted, engine.EventPassFailed, engine.EventPassProgress:
			m.pipeline.UpdatePipeline(m.pipelineData)
		}

		// Continue listening for events
		cmds = append(cmds, waitForPipelineEvent(m.events))
	}

	// Update focused pane
	cmd := m.updateFocusedPane(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Calculate pane dimensions
	paneWidth := m.width / 2
	paneHeight := (m.height - 2) / 2

	// Render top row (file tree + code view)
	topLeft := m.renderPane(m.fileTree.View(), "File Tree", paneWidth, paneHeight, m.focus == FocusPaneFileTree)
	topRight := m.renderPane(m.codeView.View(), "Code Preview", paneWidth, paneHeight, m.focus == FocusPaneCodeView)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topLeft, topRight)

	// Render bottom row (findings + pipeline)
	bottomLeft := m.renderPane(m.findings.View(), "Findings", paneWidth, paneHeight, m.focus == FocusPaneFindings)
	bottomRight := m.renderPane(m.pipeline.View(), "Pipeline", paneWidth, paneHeight, m.focus == FocusPanePipeline)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, bottomLeft, bottomRight)

	// Combine rows
	main := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

	// Add status bar
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, main, statusBar)
}

// renderPane renders a pane with borders and title
func (m Model) renderPane(content, title string, width, height int, focused bool) string {
	style := theme.PaneBorderStyle
	titleStyle := theme.MutedStyle

	if focused {
		style = theme.ActivePaneBorderStyle
		titleStyle = theme.PaneTitleStyle
	}

	styledTitle := titleStyle.Render(title)
	styledContent := style.
		Width(width - 4).
		Height(height - 3).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, styledTitle, styledContent)
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	help := theme.MutedStyle.Render("tab: cycle focus | h/j/k/l: navigate | q: quit")
	return lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Render(help)
}

// updatePaneSizes updates all pane sizes based on window size
func (m Model) updatePaneSizes() {
	paneWidth := m.width / 2
	paneHeight := (m.height - 2) / 2

	m.fileTree.SetSize(paneWidth-4, paneHeight-3)
	m.codeView.SetSize(paneWidth-4, paneHeight-3)
	m.findings.SetSize(paneWidth-4, paneHeight-3)
	m.pipeline.SetSize(paneWidth-4, paneHeight-3)
}

// updateFocusedPane updates the currently focused pane
func (m Model) updateFocusedPane(msg tea.Msg) tea.Cmd {
	switch m.focus {
	case FocusPaneFileTree:
		return m.fileTree.Update(msg)
	case FocusPaneCodeView:
		return m.codeView.Update(msg)
	case FocusPaneFindings:
		return m.findings.Update(msg)
	case FocusPanePipeline:
		return m.pipeline.Update(msg)
	}
	return nil
}

// PipelineEventMsg wraps a pipeline event
type PipelineEventMsg struct {
	Event engine.PipelineEvent
}

// waitForPipelineEvent waits for the next pipeline event
func waitForPipelineEvent(events <-chan engine.PipelineEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return nil
		}
		return PipelineEventMsg{Event: event}
	}
}
