package menu

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// MenuOption represents a menu item
type MenuOption int

const (
	MenuOptionStart MenuOption = iota
	MenuOptionModelSelect
	MenuOptionSettings
	MenuOptionExit
)

// MenuSelectionMsg is sent when a menu option is selected
type MenuSelectionMsg struct {
	Selection MenuOption
}

// BackToMenuMsg is sent when returning to the main menu
type BackToMenuMsg struct{}

// MenuModel represents the main menu
type MenuModel struct {
	projectRoot string
	selected    int
	options     []menuItem
	width       int
	height      int

	// Latest report info
	latestReport  string
	findingsCount int
	lastRunTime   time.Time
	hasReport     bool
}

type menuItem struct {
	label  string
	option MenuOption
}

// NewMenuModel creates a new menu model
func NewMenuModel(projectRoot string) *MenuModel {
	options := []menuItem{
		{label: "START ANALYSIS", option: MenuOptionStart},
		{label: "MODEL SELECT", option: MenuOptionModelSelect},
		{label: "SETTINGS", option: MenuOptionSettings},
		{label: "EXIT", option: MenuOptionExit},
	}

	m := &MenuModel{
		projectRoot: projectRoot,
		selected:    0,
		options:     options,
	}

	// Load latest report info
	m.loadReportInfo()

	return m
}

// SetSize sets the menu dimensions
func (m *MenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the menu
func (m *MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *MenuModel) Update(msg tea.Msg) (*MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.selected == int(MenuOptionExit) {
				return m, tea.Quit
			}
			// Move to exit option
			m.selected = int(MenuOptionExit)

		case "up":
			if m.selected > 0 {
				m.selected--
			}

		case "down":
			if m.selected < len(m.options)-1 {
				m.selected++
			}

		case "enter":
			// Send selection message
			selectedOption := m.options[m.selected].option
			return m, func() tea.Msg {
				return MenuSelectionMsg{Selection: selectedOption}
			}
		}
	}

	return m, nil
}

// View renders the menu
func (m *MenuModel) View() string {
	var b strings.Builder

	// Render logo
	logo := theme.RenderLogo()
	b.WriteString(centerText(logo, m.width))
	b.WriteString("\n\n")

	// Render project info
	projectName := filepath.Base(m.projectRoot)
	projectInfo := theme.MutedStyle.Render(fmt.Sprintf("Project: %s", projectName))
	b.WriteString(centerText(projectInfo, m.width))
	b.WriteString("\n")

	// Render latest report info
	if m.hasReport {
		reportInfo := theme.MutedStyle.Render(fmt.Sprintf(
			"Latest Report: %s (%d findings)",
			m.lastRunTime.Format("2006-01-02 15:04:05"),
			m.findingsCount,
		))
		b.WriteString(centerText(reportInfo, m.width))
	} else {
		reportInfo := theme.MutedStyle.Render("No reports found - run analysis to get started")
		b.WriteString(centerText(reportInfo, m.width))
	}
	b.WriteString("\n\n")

	// Render menu box
	menuContent := m.renderMenuItems()
	menuBox := m.renderMenuBox(menuContent)
	b.WriteString(centerText(menuBox, m.width))
	b.WriteString("\n\n")

	// Render help text
	helpText := theme.MutedStyle.Render("↑/↓: navigate | Enter: select | q: quit")
	b.WriteString(centerText(helpText, m.width))

	// Add padding to fill screen
	content := b.String()
	lines := strings.Split(content, "\n")
	paddingNeeded := m.height - len(lines) - 1
	if paddingNeeded > 0 {
		b.WriteString(strings.Repeat("\n", paddingNeeded))
	}

	return b.String()
}

// renderMenuItems renders the menu options
func (m *MenuModel) renderMenuItems() string {
	var items []string

	for i, item := range m.options {
		var line string

		if i == m.selected {
			// Selected item with solid background
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(theme.ColorPrimaryRed)).
				Foreground(lipgloss.Color(theme.ColorTextPrimary)).
				Bold(true).
				Padding(0, 2).
				Width(30)

			line = selectedStyle.Render("▶ " + item.label)
		} else {
			// Unselected item with dark background
			unselectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(theme.ColorBackground)).
				Foreground(lipgloss.Color(theme.ColorMuted)).
				Padding(0, 2).
				Width(30)

			line = unselectedStyle.Render("  " + item.label)
		}

		items = append(items, line)
	}

	return strings.Join(items, "\n")
}

// renderMenuBox renders the menu box with border
func (m *MenuModel) renderMenuBox(content string) string {
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.ColorPrimaryRed)).
		BorderBackground(lipgloss.Color(theme.ColorBackground)).
		Background(lipgloss.Color(theme.ColorBackground)).
		Padding(1, 0)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ColorPrimaryRed)).
		Bold(true).
		Render(" Main Menu ")

	// Combine title and content
	fullContent := title + "\n" + content

	return boxStyle.Render(fullContent)
}

// loadReportInfo loads information about the latest report
func (m *MenuModel) loadReportInfo() {
	reports, err := engine.ListReports(m.projectRoot)
	if err != nil || len(reports) == 0 {
		m.hasReport = false
		return
	}

	// Load the most recent report
	latestReport := reports[len(reports)-1]
	report, err := engine.LoadReport(latestReport)
	if err != nil {
		m.hasReport = false
		return
	}

	m.latestReport = latestReport
	m.findingsCount = len(report.Findings)
	m.lastRunTime = report.Timestamp
	m.hasReport = true
}

// centerText centers text horizontally
func centerText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var centered []string

	for _, line := range lines {
		// Remove ANSI codes for length calculation
		cleanLine := stripAnsi(line)
		lineLen := len(cleanLine)

		if lineLen >= width {
			centered = append(centered, line)
			continue
		}

		padding := (width - lineLen) / 2
		centeredLine := strings.Repeat(" ", padding) + line
		centered = append(centered, centeredLine)
	}

	return strings.Join(centered, "\n")
}

// stripAnsi removes ANSI escape codes for length calculation
func stripAnsi(str string) string {
	// Simple ANSI stripper for length calculation
	// This is a basic implementation
	var result strings.Builder
	inEscape := false

	for _, r := range str {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}
