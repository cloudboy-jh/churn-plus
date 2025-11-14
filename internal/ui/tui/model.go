package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// PaneFocus represents which pane has focus
type PaneFocus int

const (
	FocusListPane PaneFocus = iota
	FocusDetailPane
)

// BackToMenuMsg is sent when user wants to return to menu
type BackToMenuMsg struct{}

// Model is the main two-pane TUI model
type Model struct {
	projectRoot string
	config      *config.Config
	findings    []*engine.Finding

	// Panes
	listPane   *ListPane
	detailPane *DetailPane

	// State
	focus       PaneFocus
	selectedIdx int
	width       int
	height      int

	// Modal state
	showLLMModal      bool
	llmModal          *LLMModal
	showPatchPreview  bool
	patchPreviewModal *PatchPreviewModal
}

// NewModel creates a new TUI model
func NewModel(projectRoot string, findings []*engine.Finding, cfg *config.Config) *Model {
	m := &Model{
		projectRoot: projectRoot,
		config:      cfg,
		findings:    findings,
		focus:       FocusListPane,
		selectedIdx: 0,
	}

	// Create panes
	m.listPane = NewListPane(findings)
	m.detailPane = NewDetailPane()

	// Set initial selection
	if len(findings) > 0 {
		m.detailPane.SetFinding(findings[0])
	}

	return m
}

// SetSize sets the model dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate pane sizes
	leftWidth := width / 3
	rightWidth := width - leftWidth
	paneHeight := height - 2 // Reserve space for status bar

	if m.listPane != nil {
		m.listPane.SetSize(leftWidth, paneHeight)
	}
	if m.detailPane != nil {
		m.detailPane.SetSize(rightWidth, paneHeight)
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	// Handle modal updates first
	if m.showLLMModal {
		return m.updateLLMModal(msg)
	}
	if m.showPatchPreview {
		return m.updatePatchPreview(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if m.focus == FocusDetailPane {
			// Return to list pane
			m.focus = FocusListPane
			return m, nil
		}
		// From list pane, quit the TUI
		return m, tea.Quit

	case "m":
		// Return to menu
		return m, func() tea.Msg {
			return BackToMenuMsg{}
		}

	case "up":
		if m.focus == FocusListPane {
			m.navigateList(-1)
		}

	case "down":
		if m.focus == FocusListPane {
			m.navigateList(1)
		}

	case "enter":
		if m.focus == FocusListPane {
			// Switch to detail pane
			m.focus = FocusDetailPane
		}

	case "l":
		if m.focus == FocusDetailPane && len(m.findings) > 0 {
			// Send to LLM
			return m.openLLMModal()
		}

	case "p":
		if m.focus == FocusDetailPane && len(m.findings) > 0 {
			// Preview patch
			return m.openPatchPreview()
		}

	case "a":
		if m.focus == FocusDetailPane && len(m.findings) > 0 {
			// Apply patch
			return m.applyPatch()
		}
	}

	return m, nil
}

// navigateList navigates the findings list
func (m *Model) navigateList(delta int) {
	if len(m.findings) == 0 {
		return
	}

	newIdx := m.selectedIdx + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(m.findings) {
		newIdx = len(m.findings) - 1
	}

	if newIdx != m.selectedIdx {
		m.selectedIdx = newIdx
		m.listPane.SetSelected(newIdx)
		m.detailPane.SetFinding(m.findings[newIdx])
	}
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Render main two-pane layout
	mainView := m.renderMainLayout()

	// Overlay modal if active
	if m.showLLMModal && m.llmModal != nil {
		return m.renderModalOverlay(mainView, m.llmModal.View())
	}
	if m.showPatchPreview && m.patchPreviewModal != nil {
		return m.renderModalOverlay(mainView, m.patchPreviewModal.View())
	}

	return mainView
}

// renderMainLayout renders the two-pane layout
func (m *Model) renderMainLayout() string {
	// Render left pane (findings list)
	leftFocused := m.focus == FocusListPane
	leftView := m.listPane.View(leftFocused)

	// Render right pane (detail view)
	rightFocused := m.focus == FocusDetailPane
	rightView := m.detailPane.View(rightFocused)

	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)

	// Render status bar
	statusBar := m.renderStatusBar()

	// Join vertically
	return lipgloss.JoinVertical(lipgloss.Left, panes, statusBar)
}

// renderStatusBar renders the bottom status bar
func (m *Model) renderStatusBar() string {
	var helpText string

	if m.focus == FocusListPane {
		helpText = "↑/↓: navigate | Enter: select | m: menu | q: quit"
	} else {
		helpText = "l: LLM hand-off | p: preview patch | a: apply | m: menu | q: back"
	}

	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.ColorBackground)).
		Foreground(lipgloss.Color(theme.ColorMuted)).
		Width(m.width).
		Padding(0, 1)

	return statusStyle.Render(helpText)
}

// renderModalOverlay renders a modal on top of the main view
func (m *Model) renderModalOverlay(mainView, modalView string) string {
	// Calculate modal position (centered)
	mainLines := lipgloss.Height(mainView)
	modalLines := lipgloss.Height(modalView)
	modalWidth := lipgloss.Width(modalView)

	// Calculate padding to center modal
	topPadding := (mainLines - modalLines) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	leftPadding := (m.width - modalWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Create overlay by placing modal on top of main view
	overlayStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.ColorBackground))

	// Simple overlay: just render modal centered
	// For a true overlay effect, we'd need to draw the modal over the background
	// For now, we'll just center it on a dark background
	centeredModal := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalView,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color(theme.ColorBackground)),
	)

	return overlayStyle.Render(centeredModal)
}

// openLLMModal opens the LLM modal
func (m *Model) openLLMModal() (*Model, tea.Cmd) {
	if len(m.findings) == 0 {
		return m, nil
	}

	finding := m.findings[m.selectedIdx]
	m.llmModal = NewLLMModal(finding, m.config)
	m.showLLMModal = true

	return m, m.llmModal.Init()
}

// updateLLMModal updates the LLM modal
func (m *Model) updateLLMModal(msg tea.Msg) (*Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "q" || msg.String() == "esc" {
			m.showLLMModal = false
			m.llmModal = nil
			return m, nil
		}
	}

	var cmd tea.Cmd
	*m.llmModal, cmd = m.llmModal.Update(msg)

	return m, cmd
}

// openPatchPreview opens the patch preview modal
func (m *Model) openPatchPreview() (*Model, tea.Cmd) {
	if len(m.findings) == 0 {
		return m, nil
	}

	finding := m.findings[m.selectedIdx]
	m.patchPreviewModal = NewPatchPreviewModal(finding)
	m.showPatchPreview = true

	return m, nil
}

// updatePatchPreview updates the patch preview modal
func (m *Model) updatePatchPreview(msg tea.Msg) (*Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "esc":
			m.showPatchPreview = false
			m.patchPreviewModal = nil
			return m, nil
		case "a":
			// Apply patch from preview
			m.showPatchPreview = false
			m.patchPreviewModal = nil
			return m.applyPatch()
		}
	}

	return m, nil
}

// applyPatch applies the patch for the current finding
func (m *Model) applyPatch() (*Model, tea.Cmd) {
	// TODO: Implement patch application
	// For now, just a placeholder
	return m, nil
}
