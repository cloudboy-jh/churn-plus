package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
	"github.com/cloudboy-jh/churn-plus/internal/engine/providers"
	"github.com/cloudboy-jh/churn-plus/internal/theme"
)

// LLMModal handles LLM streaming interaction
type LLMModal struct {
	finding *engine.Finding
	config  *config.Config

	// State
	streaming bool
	completed bool
	response  strings.Builder
	err       error

	// Dimensions
	width  int
	height int
}

// NewLLMModal creates a new LLM modal
func NewLLMModal(finding *engine.Finding, cfg *config.Config) *LLMModal {
	return &LLMModal{
		finding:   finding,
		config:    cfg,
		streaming: false,
		completed: false,
		width:     80,
		height:    30,
	}
}

// Init initializes the modal and starts LLM streaming
func (m *LLMModal) Init() tea.Cmd {
	m.streaming = true
	return m.streamLLM()
}

// Update handles messages
func (m *LLMModal) Update(msg tea.Msg) (LLMModal, tea.Cmd) {
	switch msg := msg.(type) {
	case llmTokenMsg:
		// Append token to response
		m.response.WriteString(msg.token)
		return *m, nil

	case llmCompleteMsg:
		m.streaming = false
		m.completed = true
		return *m, nil

	case llmErrorMsg:
		m.streaming = false
		m.err = msg.err
		return *m, nil
	}

	return *m, nil
}

// View renders the modal
func (m *LLMModal) View() string {
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
	if m.streaming {
		title := theme.HighlightStyle.Render("🔄 LLM Response (streaming...)")
		content.WriteString(title)
		content.WriteString("\n\n")
	} else if m.err != nil {
		title := theme.ErrorStyle.Render("❌ LLM Error")
		content.WriteString(title)
		content.WriteString("\n\n")
	} else {
		title := theme.SuccessStyle.Render("✅ LLM Response Complete")
		content.WriteString(title)
		content.WriteString("\n\n")
	}

	// Content
	if m.err != nil {
		errorText := theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		content.WriteString(errorText)
	} else {
		// Wrap response text
		responseText := m.response.String()
		if responseText == "" {
			responseText = "Waiting for response..."
		}
		wrapped := wrapText(responseText, m.width-8)
		content.WriteString(wrapped)
	}

	// Footer
	content.WriteString("\n\n")
	if m.completed {
		footer := theme.MutedStyle.Render("Press 'q' to close | Press 'a' to apply patch")
		content.WriteString(footer)
	} else if !m.streaming {
		footer := theme.MutedStyle.Render("Press 'q' to close")
		content.WriteString(footer)
	}

	return modalStyle.Render(content.String())
}

// streamLLM starts streaming from the LLM
func (m *LLMModal) streamLLM() tea.Cmd {
	return func() tea.Msg {
		// Get model selection
		modelSelection := m.config.GetModelSelection()

		// Create provider
		var provider providers.ModelProvider
		switch modelSelection.Provider {
		case "anthropic":
			apiKey := m.config.GetAPIKey("anthropic")
			if apiKey == "" {
				return llmErrorMsg{err: fmt.Errorf("Anthropic API key not set")}
			}
			provider = providers.NewAnthropicProvider(apiKey)

		case "openai":
			apiKey := m.config.GetAPIKey("openai")
			if apiKey == "" {
				return llmErrorMsg{err: fmt.Errorf("OpenAI API key not set")}
			}
			provider = providers.NewOpenAIProvider(apiKey)

		case "google":
			apiKey := m.config.GetAPIKey("google")
			if apiKey == "" {
				return llmErrorMsg{err: fmt.Errorf("Google API key not set")}
			}
			provider = providers.NewGoogleProvider(apiKey)

		case "ollama":
			provider = providers.NewOllamaProvider("http://localhost:11434")

		default:
			return llmErrorMsg{err: fmt.Errorf("unknown provider: %s", modelSelection.Provider)}
		}

		// Build prompt
		prompt := m.buildPrompt()

		// Create request options
		opts := providers.DefaultRequestOptions()
		opts.Model = modelSelection.Model
		opts.SystemPrompt = "You are a code fixing assistant. Provide concise, actionable fixes with patches in unified diff format."

		// Stream response
		ctx := context.Background()
		tokenChan, errChan := provider.Stream(ctx, prompt, opts)

		// Read stream (this is synchronous for simplicity)
		// In a real implementation, we'd use a goroutine and send messages back
		var fullResponse strings.Builder

		for {
			select {
			case token, ok := <-tokenChan:
				if !ok {
					// Channel closed, streaming complete
					return llmCompleteMsg{}
				}
				fullResponse.WriteString(token)
				// Send token message
				return llmTokenMsg{token: token}

			case err, ok := <-errChan:
				if ok && err != nil {
					return llmErrorMsg{err: err}
				}
			}
		}
	}
}

// buildPrompt builds the prompt for the LLM
func (m *LLMModal) buildPrompt() string {
	var prompt strings.Builder

	prompt.WriteString("Fix this code issue:\n\n")
	prompt.WriteString(fmt.Sprintf("File: %s (lines %d-%d)\n",
		m.finding.File, m.finding.LineStart, m.finding.LineEnd))
	prompt.WriteString(fmt.Sprintf("Issue: %s\n", m.finding.Message))
	prompt.WriteString(fmt.Sprintf("Type: %s\n", m.finding.Kind))
	prompt.WriteString(fmt.Sprintf("Severity: %s\n\n", m.finding.Severity))

	if m.finding.Code != "" {
		prompt.WriteString("Code:\n```\n")
		prompt.WriteString(m.finding.Code)
		prompt.WriteString("\n```\n\n")
	}

	prompt.WriteString("Provide:\n")
	prompt.WriteString("1. Brief explanation of the issue\n")
	prompt.WriteString("2. A patch in unified diff format\n")
	prompt.WriteString("3. Why this fix is safe\n")

	return prompt.String()
}

// llmTokenMsg is sent when a token is received
type llmTokenMsg struct {
	token string
}

// llmCompleteMsg is sent when streaming completes
type llmCompleteMsg struct{}

// llmErrorMsg is sent when an error occurs
type llmErrorMsg struct {
	err error
}
