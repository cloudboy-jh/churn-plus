package panes

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CodeViewPane displays file content with syntax highlighting
type CodeViewPane struct {
	width      int
	height     int
	filePath   string
	lines      []string
	scroll     int
	highlights map[int]bool // Line numbers to highlight
}

// NewCodeViewPane creates a new code view pane
func NewCodeViewPane() *CodeViewPane {
	return &CodeViewPane{
		lines:      make([]string, 0),
		highlights: make(map[int]bool),
	}
}

// SetSize sets the pane dimensions
func (p *CodeViewPane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetFile loads a file for display
func (p *CodeViewPane) SetFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	p.filePath = filePath
	p.lines = make([]string, 0)
	p.scroll = 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p.lines = append(p.lines, scanner.Text())
	}

	return scanner.Err()
}

// SetHighlights sets which lines to highlight
func (p *CodeViewPane) SetHighlights(lines map[int]bool) {
	p.highlights = lines
}

// JumpToLine scrolls to a specific line
func (p *CodeViewPane) JumpToLine(line int) {
	if line < 1 || line > len(p.lines) {
		return
	}

	// Center the line in view
	p.scroll = line - p.height/2
	if p.scroll < 0 {
		p.scroll = 0
	}
	if p.scroll > len(p.lines)-p.height {
		p.scroll = len(p.lines) - p.height
	}
}

// Update handles messages
func (p *CodeViewPane) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if p.scroll < len(p.lines)-p.height {
				p.scroll++
			}
		case "k", "up":
			if p.scroll > 0 {
				p.scroll--
			}
		case "g":
			p.scroll = 0
		case "G":
			p.scroll = len(p.lines) - p.height
			if p.scroll < 0 {
				p.scroll = 0
			}
		}
	}
	return nil
}

// View renders the pane
func (p *CodeViewPane) View() string {
	if len(p.lines) == 0 {
		return "No file selected"
	}

	var sb strings.Builder

	// Display visible lines
	start := p.scroll
	end := p.scroll + p.height
	if end > len(p.lines) {
		end = len(p.lines)
	}

	for i := start; i < end; i++ {
		lineNum := i + 1
		marker := " "
		if p.highlights[lineNum] {
			marker = "►"
		}

		line := fmt.Sprintf("%s%4d │ %s\n", marker, lineNum, p.lines[i])

		// Truncate if too long
		if len(line) > p.width {
			line = line[:p.width-3] + "...\n"
		}

		sb.WriteString(line)
	}

	return sb.String()
}
