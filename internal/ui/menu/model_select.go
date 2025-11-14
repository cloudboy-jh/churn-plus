package menu

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine/providers"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// ModelSelectStep represents the current step in model selection
type ModelSelectStep int

const (
	StepProvider ModelSelectStep = iota
	StepModel
)

// ModelSelectModel handles model selection
type ModelSelectModel struct {
	config   *config.Config
	step     ModelSelectStep
	selected int
	width    int
	height   int

	// Provider selection
	providers []providerOption

	// Model selection
	models           []string
	selectedProvider string
	loadingModels    bool
}

type providerOption struct {
	name  string
	label string
}

// NewModelSelectModel creates a new model selection model
func NewModelSelectModel(cfg *config.Config) *ModelSelectModel {
	providers := []providerOption{
		{name: "anthropic", label: "Anthropic (Claude)"},
		{name: "openai", label: "OpenAI (GPT)"},
		{name: "google", label: "Google (Gemini)"},
		{name: "ollama", label: "Ollama (Local)"},
	}

	return &ModelSelectModel{
		config:    cfg,
		step:      StepProvider,
		selected:  0,
		providers: providers,
	}
}

// SetSize sets the model dimensions
func (m *ModelSelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the model
func (m *ModelSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *ModelSelectModel) Update(msg tea.Msg) (*ModelSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Go back to main menu
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}

		case "up":
			if m.selected > 0 {
				m.selected--
			}

		case "down":
			maxItems := 0
			if m.step == StepProvider {
				maxItems = len(m.providers)
			} else {
				maxItems = len(m.models) + 1 // +1 for "Back" option
			}

			if m.selected < maxItems-1 {
				m.selected++
			}

		case "enter":
			return m.handleSelection()
		}

	case modelsLoadedMsg:
		m.models = msg.models
		m.loadingModels = false
		return m, nil
	}

	return m, nil
}

// View renders the model selection
func (m *ModelSelectModel) View() string {
	var b strings.Builder

	// Render title
	b.WriteString("\n\n")
	title := theme.TitleStyle.Render("MODEL SELECTION")
	b.WriteString(centerText(title, m.width))
	b.WriteString("\n\n")

	// Render current step
	var content string
	if m.step == StepProvider {
		content = m.renderProviderSelection()
	} else {
		if m.loadingModels {
			content = m.renderLoading()
		} else {
			content = m.renderModelSelection()
		}
	}

	menuBox := m.renderBox(content)
	b.WriteString(centerText(menuBox, m.width))
	b.WriteString("\n\n")

	// Render help text
	helpText := theme.MutedStyle.Render("↑/↓: navigate | Enter: select | q: back to menu")
	b.WriteString(centerText(helpText, m.width))

	return b.String()
}

// renderProviderSelection renders the provider selection step
func (m *ModelSelectModel) renderProviderSelection() string {
	var items []string

	for i, provider := range m.providers {
		var line string

		if i == m.selected {
			// Selected with solid background
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(theme.ColorPrimaryRed)).
				Foreground(lipgloss.Color(theme.ColorTextPrimary)).
				Bold(true).
				Padding(0, 2).
				Width(35)

			line = selectedStyle.Render("▶ " + provider.label)
		} else {
			// Unselected with dark background
			unselectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(theme.ColorBackground)).
				Foreground(lipgloss.Color(theme.ColorMuted)).
				Padding(0, 2).
				Width(35)

			line = unselectedStyle.Render("  " + provider.label)
		}

		items = append(items, line)
	}

	// Add back option
	backStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.ColorBackground)).
		Foreground(lipgloss.Color(theme.ColorMuted)).
		Padding(0, 2).
		Width(35)
	items = append(items, backStyle.Render("  < Back to Menu"))

	return strings.Join(items, "\n")
}

// renderModelSelection renders the model selection step
func (m *ModelSelectModel) renderModelSelection() string {
	var items []string

	for i, model := range m.models {
		var line string

		if i == m.selected {
			// Selected with solid background
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(theme.ColorPrimaryRed)).
				Foreground(lipgloss.Color(theme.ColorTextPrimary)).
				Bold(true).
				Padding(0, 2).
				Width(40)

			line = selectedStyle.Render("▶ " + model)
		} else {
			// Unselected with dark background
			unselectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(theme.ColorBackground)).
				Foreground(lipgloss.Color(theme.ColorMuted)).
				Padding(0, 2).
				Width(40)

			line = unselectedStyle.Render("  " + model)
		}

		items = append(items, line)
	}

	// Add back option
	backLabel := "< Back to Providers"
	if m.selected == len(m.models) {
		selectedStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.ColorPrimaryRed)).
			Foreground(lipgloss.Color(theme.ColorTextPrimary)).
			Bold(true).
			Padding(0, 2).
			Width(40)
		items = append(items, selectedStyle.Render("▶ "+backLabel))
	} else {
		backStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.ColorBackground)).
			Foreground(lipgloss.Color(theme.ColorMuted)).
			Padding(0, 2).
			Width(40)
		items = append(items, backStyle.Render("  "+backLabel))
	}

	return strings.Join(items, "\n")
}

// renderLoading renders a loading message
func (m *ModelSelectModel) renderLoading() string {
	loadingStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.ColorBackground)).
		Foreground(lipgloss.Color(theme.ColorMuted)).
		Padding(2, 4).
		Width(40)

	return loadingStyle.Render("Loading models...")
}

// renderBox renders content in a box
func (m *ModelSelectModel) renderBox(content string) string {
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.ColorPrimaryRed)).
		BorderBackground(lipgloss.Color(theme.ColorBackground)).
		Background(lipgloss.Color(theme.ColorBackground)).
		Padding(1, 0)

	var title string
	if m.step == StepProvider {
		title = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
			Bold(true).
			Render(" Select Provider ")
	} else {
		title = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
			Bold(true).
			Render(" Select Model ")
	}

	fullContent := title + "\n" + content

	return boxStyle.Render(fullContent)
}

// handleSelection processes the current selection
func (m *ModelSelectModel) handleSelection() (*ModelSelectModel, tea.Cmd) {
	if m.step == StepProvider {
		// Check if "Back" was selected
		if m.selected >= len(m.providers) {
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		}

		// Move to model selection
		m.selectedProvider = m.providers[m.selected].name
		m.step = StepModel
		m.selected = 0
		m.loadingModels = true

		// Load models for selected provider
		return m, m.loadModels()

	} else {
		// Model selection step
		if m.selected >= len(m.models) {
			// Back to provider selection
			m.step = StepProvider
			m.selected = 0
			m.models = nil
			return m, nil
		}

		// Save selected model to config
		selectedModel := m.models[m.selected]
		m.config.Project.Model = config.ModelSelection{
			Provider: m.selectedProvider,
			Model:    selectedModel,
		}

		// TODO: Save config to disk

		// Return to menu
		return m, func() tea.Msg {
			return BackToMenuMsg{}
		}
	}
}

// loadModels loads available models for the selected provider
func (m *ModelSelectModel) loadModels() tea.Cmd {
	return func() tea.Msg {
		var provider providers.ModelProvider

		switch m.selectedProvider {
		case "anthropic":
			apiKey := m.config.GetAPIKey("anthropic")
			provider = providers.NewAnthropicProvider(apiKey)
		case "openai":
			apiKey := m.config.GetAPIKey("openai")
			provider = providers.NewOpenAIProvider(apiKey)
		case "google":
			apiKey := m.config.GetAPIKey("google")
			provider = providers.NewGoogleProvider(apiKey)
		case "ollama":
			provider = providers.NewOllamaProvider("http://localhost:11434")
		default:
			return modelsLoadedMsg{models: []string{}}
		}

		models, err := provider.ListModels(context.Background())
		if err != nil {
			// Return empty list on error
			return modelsLoadedMsg{models: []string{}}
		}

		return modelsLoadedMsg{models: models}
	}
}

// modelsLoadedMsg is sent when models are loaded
type modelsLoadedMsg struct {
	models []string
}
