package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// MenuModel represents the interactive start menu
type MenuModel struct {
	width  int
	height int

	// Menu state
	selectedIndex int
	menuItems     []string

	// Data
	cfg     *config.Config
	context *engine.ProjectContext

	// Submenu
	inSubmenu       bool
	submenuType     SubmenuType
	pipelineSubmenu PipelineSubmenuModel
	settingsSubmenu SettingsSubmenuModel
}

// SettingsSubmenuModel handles settings configuration
type SettingsSubmenuModel struct {
	selectedIndex int
	settingItems  []string
}

// SubmenuType defines the type of submenu
type SubmenuType int

const (
	SubmenuNone SubmenuType = iota
	SubmenuModelPipeline
	SubmenuSettings
)

// PipelineSubmenuModel handles pipeline configuration
type PipelineSubmenuModel struct {
	selectedIndex int
	passes        []config.PassConfig
	editing       bool
	editField     int // 0=name, 1=model, 2=provider, 3=enabled
}

// NewMenuModel creates a new menu model
func NewMenuModel(cfg *config.Config, ctx *engine.ProjectContext) MenuModel {
	// Initialize pipeline submenu with current or default passes
	passes := getDefaultPasses(cfg)
	if cfg.Project.Pipeline != nil && len(cfg.Project.Pipeline.Passes) > 0 {
		passes = cfg.Project.Pipeline.Passes
	}

	return MenuModel{
		selectedIndex: 0,
		menuItems: []string{
			"Start Analysis",
			"Configure Model Pipeline",
			"Settings",
			"Exit",
		},
		cfg:         cfg,
		context:     ctx,
		inSubmenu:   false,
		submenuType: SubmenuNone,
		pipelineSubmenu: PipelineSubmenuModel{
			selectedIndex: 0,
			passes:        passes,
			editing:       false,
		},
		settingsSubmenu: SettingsSubmenuModel{
			selectedIndex: 0,
			settingItems: []string{
				"API Keys",
				"Default Model",
				"Concurrency Limits",
				"Cache Settings",
				"UI Settings",
			},
		},
	}
}

// getDefaultPasses returns default pipeline configuration
func getDefaultPasses(cfg *config.Config) []config.PassConfig {
	modelSelection := cfg.GetModelSelection()
	lintModel := "claude-3-5-haiku-20241022"
	if modelSelection.Provider == "openai" {
		lintModel = "gpt-3.5-turbo"
	}

	return []config.PassConfig{
		{
			Name:        "lint",
			Description: "Quick structural checks for unused code and basic issues",
			Enabled:     true,
			Model:       lintModel,
			Provider:    modelSelection.Provider,
		},
		{
			Name:        "refactor",
			Description: "Deep analysis for architectural improvements",
			Enabled:     true,
			Model:       modelSelection.Model,
			Provider:    modelSelection.Provider,
		},
		{
			Name:        "summary",
			Description: "Coherence check and overall assessment",
			Enabled:     true,
			Model:       modelSelection.Model,
			Provider:    modelSelection.Provider,
		},
	}
}

// Init initializes the menu
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.inSubmenu {
			// Handle submenu navigation
			return m.updateSubmenu(msg)
		}

		// Main menu navigation
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.menuItems)-1 {
				m.selectedIndex++
			}

		case "enter":
			return m.handleSelection()
		}

	case StartAnalysisMsg:
		// Signal to parent to switch to TUI mode
		return m, func() tea.Msg { return msg }
	}

	return m, nil
}

// handleSelection handles menu item selection
func (m MenuModel) handleSelection() (tea.Model, tea.Cmd) {
	switch m.selectedIndex {
	case 0: // Start Analysis
		return m, func() tea.Msg {
			return StartAnalysisMsg{}
		}

	case 1: // Configure Model Pipeline
		m.inSubmenu = true
		m.submenuType = SubmenuModelPipeline
		return m, nil

	case 2: // Settings
		m.inSubmenu = true
		m.submenuType = SubmenuSettings
		return m, nil

	case 3: // Exit
		return m, tea.Quit
	}

	return m, nil
}

// updateSubmenu handles submenu updates
func (m MenuModel) updateSubmenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.submenuType {
	case SubmenuModelPipeline:
		return m.updatePipelineSubmenu(msg)
	case SubmenuSettings:
		return m.updateSettingsSubmenu(msg)
	}
	return m, nil
}

// updatePipelineSubmenu handles pipeline submenu navigation
func (m MenuModel) updatePipelineSubmenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inSubmenu = false
		m.submenuType = SubmenuNone
		return m, nil

	case "up", "k":
		if m.pipelineSubmenu.selectedIndex > 0 {
			m.pipelineSubmenu.selectedIndex--
		}

	case "down", "j":
		if m.pipelineSubmenu.selectedIndex < len(m.pipelineSubmenu.passes) {
			m.pipelineSubmenu.selectedIndex++
		}

	case "enter", " ":
		// Toggle enabled/disabled for the selected pass
		if m.pipelineSubmenu.selectedIndex < len(m.pipelineSubmenu.passes) {
			pass := &m.pipelineSubmenu.passes[m.pipelineSubmenu.selectedIndex]
			pass.Enabled = !pass.Enabled
		} else if m.pipelineSubmenu.selectedIndex == len(m.pipelineSubmenu.passes) {
			// Save configuration
			return m.savePipelineConfig()
		}

	case "a":
		// Add new pass
		m.pipelineSubmenu.passes = append(m.pipelineSubmenu.passes, config.PassConfig{
			Name:        "new-pass",
			Description: "New pass description",
			Enabled:     true,
			Model:       m.cfg.GetModelSelection().Model,
			Provider:    m.cfg.GetModelSelection().Provider,
		})
	}

	return m, nil
}

// updateSettingsSubmenu handles settings submenu navigation
func (m MenuModel) updateSettingsSubmenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inSubmenu = false
		m.submenuType = SubmenuNone
		return m, nil

	case "up", "k":
		if m.settingsSubmenu.selectedIndex > 0 {
			m.settingsSubmenu.selectedIndex--
		}

	case "down", "j":
		if m.settingsSubmenu.selectedIndex < len(m.settingsSubmenu.settingItems)-1 {
			m.settingsSubmenu.selectedIndex++
		}

	case "enter":
		// Handle settings selection
		// For now, just placeholder
	}

	return m, nil
}

// savePipelineConfig saves the pipeline configuration
func (m MenuModel) savePipelineConfig() (tea.Model, tea.Cmd) {
	// Update config
	if m.cfg.Project.Pipeline == nil {
		m.cfg.Project.Pipeline = &config.PipelineConfig{}
	}
	m.cfg.Project.Pipeline.Passes = m.pipelineSubmenu.passes

	// Save to file
	if err := config.SaveProjectConfig(m.context.RootPath, m.cfg.Project); err != nil {
		// Handle error - for now just return
		return m, nil
	}

	// Go back to main menu
	m.inSubmenu = false
	m.submenuType = SubmenuNone

	return m, nil
}

// View renders the menu
func (m MenuModel) View() string {
	if m.inSubmenu {
		return m.renderSubmenu()
	}

	var s strings.Builder

	// Logo
	s.WriteString(theme.RenderLogo())
	s.WriteString("\n\n")

	// Project info
	s.WriteString(m.renderProjectInfo())
	s.WriteString("\n\n")

	// Current pipeline
	s.WriteString(m.renderCurrentPipeline())
	s.WriteString("\n\n")

	// Menu
	s.WriteString(m.renderMenu())
	s.WriteString("\n\n")

	// Help
	s.WriteString(theme.MutedStyle.Render("Use ↑/↓ arrows to navigate, ENTER to select, q/ESC to quit"))

	return s.String()
}

// renderProjectInfo renders project information
func (m MenuModel) renderProjectInfo() string {
	return fmt.Sprintf("%s: %s\n%s: %d | %s: %s | %s: %s",
		theme.HighlightStyle.Render("Project"),
		m.context.RootPath,
		theme.MutedStyle.Render("Files"),
		m.context.FileCount,
		theme.MutedStyle.Render("Languages"),
		strings.Join(m.context.Languages, ", "),
		theme.MutedStyle.Render("Frameworks"),
		strings.Join(m.context.Frameworks, ", "),
	)
}

// renderCurrentPipeline shows the current model configuration
func (m MenuModel) renderCurrentPipeline() string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render("Current Model Pipeline"))
	s.WriteString("\n")

	modelSelection := m.cfg.GetModelSelection()

	// For now, show default configuration
	// TODO: Read from pipeline config once implemented
	s.WriteString(fmt.Sprintf("  1. Lint/Sanity: %s (%s)\n", "claude-3.5-haiku", modelSelection.Provider))
	s.WriteString(fmt.Sprintf("  2. Refactor: %s (%s)\n", modelSelection.Model, modelSelection.Provider))
	s.WriteString(fmt.Sprintf("  3. Summary: %s (%s)\n", modelSelection.Model, modelSelection.Provider))

	return s.String()
}

// renderMenu renders the menu items
func (m MenuModel) renderMenu() string {
	var s strings.Builder

	// Top border
	s.WriteString("┌─ " + theme.TitleStyle.Render("Menu") + " ──────────────────────────────┐\n")

	// Menu items
	for i, item := range m.menuItems {
		prefix := "  "
		suffix := ""

		if i == m.selectedIndex {
			prefix = theme.HighlightStyle.Render("> ")
		}

		if i == 0 {
			suffix = theme.MutedStyle.Render(" ENTER")
		} else if i == 3 {
			suffix = theme.MutedStyle.Render(" ESC")
		}

		line := fmt.Sprintf("│ %s%-30s%s │\n", prefix, item, suffix)
		s.WriteString(line)
	}

	// Bottom border
	s.WriteString("└─────────────────────────────────────┘")

	return s.String()
}

// renderSubmenu renders the active submenu
func (m MenuModel) renderSubmenu() string {
	switch m.submenuType {
	case SubmenuModelPipeline:
		return m.renderModelPipelineSubmenu()
	case SubmenuSettings:
		return m.renderSettingsSubmenu()
	default:
		return "Unknown submenu"
	}
}

// renderModelPipelineSubmenu renders the model pipeline configuration
func (m MenuModel) renderModelPipelineSubmenu() string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render("Configure Model Pipeline"))
	s.WriteString("\n\n")

	s.WriteString("Configure the analysis passes for your project.\n")
	s.WriteString("Use SPACE/ENTER to toggle pass enabled/disabled.\n\n")

	// Render passes
	s.WriteString("┌─ Passes ────────────────────────────────────────────┐\n")
	for i, pass := range m.pipelineSubmenu.passes {
		prefix := "  "
		if i == m.pipelineSubmenu.selectedIndex {
			prefix = theme.HighlightStyle.Render("> ")
		}

		status := "[ ]"
		if pass.Enabled {
			status = theme.HighlightStyle.Render("[✓]")
		}

		line := fmt.Sprintf("│ %s%s %-20s │\n", prefix, status, pass.Name)
		s.WriteString(line)

		// Show details for selected pass
		if i == m.pipelineSubmenu.selectedIndex {
			s.WriteString(fmt.Sprintf("│     %s\n", theme.MutedStyle.Render(pass.Description)))
			s.WriteString(fmt.Sprintf("│     Model: %s (%s)\n", pass.Model, pass.Provider))
		}
	}

	// Save option
	prefix := "  "
	if m.pipelineSubmenu.selectedIndex == len(m.pipelineSubmenu.passes) {
		prefix = theme.HighlightStyle.Render("> ")
	}
	s.WriteString(fmt.Sprintf("│ %s%-38s │\n", prefix, "[Save Configuration]"))
	s.WriteString("└─────────────────────────────────────────────────────┘\n")

	s.WriteString("\n")
	s.WriteString(theme.MutedStyle.Render("↑/↓: Navigate | SPACE/ENTER: Toggle/Save | A: Add pass | ESC: Back"))

	return s.String()
}

// renderSettingsSubmenu renders the settings menu
func (m MenuModel) renderSettingsSubmenu() string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render("Settings"))
	s.WriteString("\n\n")

	s.WriteString("Configure global settings for churn-plus.\n\n")

	// Render settings items
	s.WriteString("┌─ Configuration ─────────────────────────────────────┐\n")
	for i, item := range m.settingsSubmenu.settingItems {
		prefix := "  "
		if i == m.settingsSubmenu.selectedIndex {
			prefix = theme.HighlightStyle.Render("> ")
		}

		line := fmt.Sprintf("│ %s%-45s │\n", prefix, item)
		s.WriteString(line)

		// Show current values for selected item
		if i == m.settingsSubmenu.selectedIndex {
			switch item {
			case "API Keys":
				hasAnthropic := m.cfg.Global.APIKeys.Anthropic != ""
				hasOpenAI := m.cfg.Global.APIKeys.OpenAI != ""
				hasGoogle := m.cfg.Global.APIKeys.Google != ""
				s.WriteString(fmt.Sprintf("│     Anthropic: %s | OpenAI: %s | Google: %s\n",
					formatKeyStatus(hasAnthropic),
					formatKeyStatus(hasOpenAI),
					formatKeyStatus(hasGoogle)))
			case "Default Model":
				s.WriteString(fmt.Sprintf("│     %s (%s)\n",
					m.cfg.Global.DefaultModel.Model,
					m.cfg.Global.DefaultModel.Provider))
			case "Concurrency Limits":
				s.WriteString(fmt.Sprintf("│     Anthropic: %d | OpenAI: %d | Google: %d | Ollama: %d\n",
					m.cfg.Global.Concurrency.Anthropic,
					m.cfg.Global.Concurrency.OpenAI,
					m.cfg.Global.Concurrency.Google,
					m.cfg.Global.Concurrency.Ollama))
			case "Cache Settings":
				s.WriteString(fmt.Sprintf("│     Enabled: %v | TTL: %dh | Max Size: %dMB\n",
					m.cfg.Global.Cache.Enabled,
					m.cfg.Global.Cache.TTL,
					m.cfg.Global.Cache.MaxSize))
			case "UI Settings":
				s.WriteString(fmt.Sprintf("│     Theme: %s | Line Numbers: %v | Syntax: %v\n",
					m.cfg.Global.UI.Theme,
					m.cfg.Global.UI.ShowLineNumbers,
					m.cfg.Global.UI.SyntaxHighlight))
			}
		}
	}
	s.WriteString("└─────────────────────────────────────────────────────┘\n")

	s.WriteString("\n")
	s.WriteString(theme.MutedStyle.Render("↑/↓: Navigate | ENTER: Edit | ESC: Back"))
	s.WriteString("\n\n")
	s.WriteString(theme.MutedStyle.Render("Note: API keys can be configured via environment variables or ~/.churn/config.json"))

	return s.String()
}

// formatKeyStatus formats API key status
func formatKeyStatus(hasKey bool) string {
	if hasKey {
		return theme.HighlightStyle.Render("✓ Set")
	}
	return theme.MutedStyle.Render("✗ Not set")
}

// StartAnalysisMsg signals to start the analysis
type StartAnalysisMsg struct{}
