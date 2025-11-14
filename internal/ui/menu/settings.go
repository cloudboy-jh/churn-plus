package menu

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// SettingsModel displays current configuration
type SettingsModel struct {
	config      *config.Config
	projectRoot string
	width       int
	height      int
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(cfg *config.Config, projectRoot string) *SettingsModel {
	return &SettingsModel{
		config:      cfg,
		projectRoot: projectRoot,
	}
}

// SetSize sets the settings dimensions
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the settings
func (m *SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *SettingsModel) Update(msg tea.Msg) (*SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter":
			// Return to main menu
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		}
	}

	return m, nil
}

// View renders the settings
func (m *SettingsModel) View() string {
	var b strings.Builder

	// Render title
	b.WriteString("\n\n")
	title := theme.TitleStyle.Render("SETTINGS")
	b.WriteString(centerText(title, m.width))
	b.WriteString("\n\n")

	// Render settings content
	content := m.renderSettings()
	settingsBox := m.renderBox(content)
	b.WriteString(centerText(settingsBox, m.width))
	b.WriteString("\n\n")

	// Render help text
	helpText := theme.MutedStyle.Render("Press 'q' or Enter to go back to menu")
	b.WriteString(centerText(helpText, m.width))

	return b.String()
}

// renderSettings renders the settings content
func (m *SettingsModel) renderSettings() string {
	var items []string

	// Get active model selection
	modelSelection := m.config.GetModelSelection()

	// Style for labels
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
		Bold(true)

	// Style for values
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorTextPrimary))

	// Style for sensitive values
	sensitiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorMuted))

	// Provider and model
	items = append(items, labelStyle.Render("Provider: ")+valueStyle.Render(modelSelection.Provider))
	items = append(items, labelStyle.Render("Model: ")+valueStyle.Render(modelSelection.Model))
	items = append(items, "")

	// API Keys
	items = append(items, labelStyle.Render("API Keys:"))

	anthropicKey := m.config.Global.APIKeys.Anthropic
	if anthropicKey != "" {
		maskedKey := maskAPIKey(anthropicKey)
		items = append(items, "  Anthropic: "+sensitiveStyle.Render(maskedKey))
	} else {
		items = append(items, "  Anthropic: "+theme.MutedStyle.Render("not set"))
	}

	openaiKey := m.config.Global.APIKeys.OpenAI
	if openaiKey != "" {
		maskedKey := maskAPIKey(openaiKey)
		items = append(items, "  OpenAI:    "+sensitiveStyle.Render(maskedKey))
	} else {
		items = append(items, "  OpenAI:    "+theme.MutedStyle.Render("not set"))
	}

	googleKey := m.config.Global.APIKeys.Google
	if googleKey != "" {
		maskedKey := maskAPIKey(googleKey)
		items = append(items, "  Google:    "+sensitiveStyle.Render(maskedKey))
	} else {
		items = append(items, "  Google:    "+theme.MutedStyle.Render("not set"))
	}

	items = append(items, "")

	// Concurrency settings
	items = append(items, labelStyle.Render("Concurrency Limits:"))
	items = append(items, fmt.Sprintf("  Anthropic: %s", valueStyle.Render(fmt.Sprintf("%d", m.config.Global.Concurrency.Anthropic))))
	items = append(items, fmt.Sprintf("  OpenAI:    %s", valueStyle.Render(fmt.Sprintf("%d", m.config.Global.Concurrency.OpenAI))))
	items = append(items, fmt.Sprintf("  Google:    %s", valueStyle.Render(fmt.Sprintf("%d", m.config.Global.Concurrency.Google))))
	items = append(items, fmt.Sprintf("  Ollama:    %s", valueStyle.Render(fmt.Sprintf("%d", m.config.Global.Concurrency.Ollama))))
	items = append(items, "")

	// Cache settings
	items = append(items, labelStyle.Render("Cache:"))
	cacheEnabled := "disabled"
	if m.config.Global.Cache.Enabled {
		cacheEnabled = "enabled"
	}
	items = append(items, "  Status: "+valueStyle.Render(cacheEnabled))
	items = append(items, fmt.Sprintf("  TTL:    %s", valueStyle.Render(fmt.Sprintf("%d hours", m.config.Global.Cache.TTL))))
	items = append(items, fmt.Sprintf("  Size:   %s", valueStyle.Render(fmt.Sprintf("%d MB", m.config.Global.Cache.MaxSize))))
	items = append(items, "")

	// Config file locations
	items = append(items, labelStyle.Render("Configuration Files:"))

	globalConfigPath, _ := config.GetGlobalConfigPath()
	items = append(items, "  Global:  "+theme.MutedStyle.Render(globalConfigPath))

	projectConfigPath := config.GetProjectConfigPath(m.projectRoot)
	items = append(items, "  Project: "+theme.MutedStyle.Render(projectConfigPath))

	// Wrap content in background style
	contentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.ColorBackground)).
		Foreground(lipgloss.Color(theme.ColorTextPrimary)).
		Padding(0, 2)

	return contentStyle.Render(strings.Join(items, "\n"))
}

// renderBox renders content in a box
func (m *SettingsModel) renderBox(content string) string {
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.ColorPrimaryRed)).
		BorderBackground(lipgloss.Color(theme.ColorBackground)).
		Background(lipgloss.Color(theme.ColorBackground)).
		Padding(1, 0).
		Width(70)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
		Bold(true).
		Render(" Current Configuration ")

	fullContent := title + "\n" + content

	return boxStyle.Render(fullContent)
}

// maskAPIKey masks an API key for display
func maskAPIKey(key string) string {
	if len(key) <= 10 {
		return strings.Repeat("*", len(key))
	}

	// Show first 7 characters and mask the rest
	prefix := key[:7]
	masked := strings.Repeat("*", len(key)-7)
	return prefix + masked
}
