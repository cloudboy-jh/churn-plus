package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/ui/menu"
	"github.com/cloudboy-jh/churn-plus/internal/ui/tui"
)

// AppState represents the current state of the application
type AppState int

const (
	StateMenu AppState = iota
	StateModelSelect
	StateSettings
	StateTUI
	StateLLMModal
	StatePatchPreview
	StateConfirmation
)

// AppModel is the root BubbleTea model
type AppModel struct {
	state       AppState
	projectRoot string
	config      *config.Config

	// Sub-models for different states
	menuModel        *menu.MenuModel
	modelSelectModel *menu.ModelSelectModel
	settingsModel    *menu.SettingsModel
	tuiModel         *tui.Model

	// Window dimensions
	width  int
	height int

	// Error handling
	err error
}

// NewAppModel creates a new application model
func NewAppModel(projectRoot string) AppModel {
	// Load configuration
	cfg, err := config.Load(projectRoot)
	if err != nil {
		// If config fails to load, create default
		cfg = &config.Config{
			Global:  config.DefaultGlobalConfig(),
			Project: config.DefaultProjectConfig(),
		}
	}

	return AppModel{
		state:       StateMenu,
		projectRoot: projectRoot,
		config:      cfg,
		menuModel:   menu.NewMenuModel(projectRoot),
		err:         nil,
	}
}

// Init initializes the model
func (m AppModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and state transitions
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update all sub-models with window size
		if m.menuModel != nil {
			m.menuModel.SetSize(msg.Width, msg.Height)
		}
		if m.tuiModel != nil {
			m.tuiModel.SetSize(msg.Width, msg.Height)
		}

		return m, nil

	case tea.KeyMsg:
		// Global quit handler
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case menu.MenuSelectionMsg:
		// Handle menu selection
		return m.handleMenuSelection(msg)

	case menu.BackToMenuMsg:
		// Return to main menu
		m.state = StateMenu
		return m, nil

	case tui.BackToMenuMsg:
		// Return to main menu from TUI
		m.state = StateMenu
		return m, nil
	}

	// Delegate to current state's sub-model
	return m.updateCurrentState(msg)
}

// View renders the current state
func (m AppModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress Ctrl+C to quit", m.err)
	}

	switch m.state {
	case StateMenu:
		if m.menuModel != nil {
			return m.menuModel.View()
		}
		return "Loading menu..."

	case StateModelSelect:
		if m.modelSelectModel != nil {
			return m.modelSelectModel.View()
		}
		return "Loading model selection..."

	case StateSettings:
		if m.settingsModel != nil {
			return m.settingsModel.View()
		}
		return "Loading settings..."

	case StateTUI:
		if m.tuiModel != nil {
			return m.tuiModel.View()
		}
		return "Loading TUI..."

	default:
		return "Unknown state"
	}
}

// handleMenuSelection processes menu selections and transitions states
func (m AppModel) handleMenuSelection(msg menu.MenuSelectionMsg) (AppModel, tea.Cmd) {
	switch msg.Selection {
	case menu.MenuOptionStart:
		// Load findings and transition to TUI
		findings, err := m.loadFindings()
		if err != nil {
			m.err = fmt.Errorf("failed to load findings: %w", err)
			return m, nil
		}

		// Create TUI model
		m.tuiModel = tui.NewModel(m.projectRoot, findings, m.config)
		m.tuiModel.SetSize(m.width, m.height)
		m.state = StateTUI

		return m, m.tuiModel.Init()

	case menu.MenuOptionModelSelect:
		// Create model select model
		m.modelSelectModel = menu.NewModelSelectModel(m.config)
		m.modelSelectModel.SetSize(m.width, m.height)
		m.state = StateModelSelect
		return m, nil

	case menu.MenuOptionSettings:
		// Create settings model
		m.settingsModel = menu.NewSettingsModel(m.config, m.projectRoot)
		m.settingsModel.SetSize(m.width, m.height)
		m.state = StateSettings
		return m, nil

	case menu.MenuOptionExit:
		return m, tea.Quit
	}

	return m, nil
}

// updateCurrentState delegates update to the current state's sub-model
func (m AppModel) updateCurrentState(msg tea.Msg) (AppModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case StateMenu:
		if m.menuModel != nil {
			updatedMenu, menuCmd := m.menuModel.Update(msg)
			m.menuModel = updatedMenu
			cmd = menuCmd
		}

	case StateModelSelect:
		if m.modelSelectModel != nil {
			updatedModelSelect, msCmd := m.modelSelectModel.Update(msg)
			m.modelSelectModel = updatedModelSelect
			cmd = msCmd
		}

	case StateSettings:
		if m.settingsModel != nil {
			updatedSettings, sCmd := m.settingsModel.Update(msg)
			m.settingsModel = updatedSettings
			cmd = sCmd
		}

	case StateTUI:
		if m.tuiModel != nil {
			updatedTUI, tCmd := m.tuiModel.Update(msg)
			m.tuiModel = updatedTUI
			cmd = tCmd
		}
	}

	return m, cmd
}

// loadFindings loads findings from the most recent report
func (m AppModel) loadFindings() ([]*engine.Finding, error) {
	// List all reports
	reports, err := engine.ListReports(m.projectRoot)
	if err != nil {
		return nil, err
	}

	if len(reports) == 0 {
		// No reports found, return empty slice
		return []*engine.Finding{}, nil
	}

	// Load the most recent report (last in list)
	latestReport := reports[len(reports)-1]
	report, err := engine.LoadReport(latestReport)
	if err != nil {
		return nil, err
	}

	return report.Findings, nil
}
