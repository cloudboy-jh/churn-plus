package panes

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudboy-jh/churn-plus/internal/engine"
)

// FileTreePane displays the project file tree
type FileTreePane struct {
	width    int
	height   int
	tree     *engine.FileNode
	selected int
	scroll   int
	items    []string
}

// NewFileTreePane creates a new file tree pane
func NewFileTreePane() *FileTreePane {
	return &FileTreePane{
		items: make([]string, 0),
	}
}

// SetSize sets the pane dimensions
func (p *FileTreePane) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetTree sets the file tree data
func (p *FileTreePane) SetTree(tree *engine.FileNode) {
	p.tree = tree
	p.items = p.flattenTree(tree, 0)
}

// Update handles messages
func (p *FileTreePane) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if p.selected < len(p.items)-1 {
				p.selected++
				if p.selected >= p.scroll+p.height {
					p.scroll++
				}
			}
		case "k", "up":
			if p.selected > 0 {
				p.selected--
				if p.selected < p.scroll {
					p.scroll--
				}
			}
		}
	}
	return nil
}

// View renders the pane
func (p *FileTreePane) View() string {
	if len(p.items) == 0 {
		return "No files to display"
	}

	var sb strings.Builder

	// Display visible items
	start := p.scroll
	end := p.scroll + p.height
	if end > len(p.items) {
		end = len(p.items)
	}

	for i := start; i < end; i++ {
		line := p.items[i]
		if i == p.selected {
			sb.WriteString("▶ " + line + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}

	return sb.String()
}

// flattenTree converts tree structure to flat list for display
func (p *FileTreePane) flattenTree(node *engine.FileNode, depth int) []string {
	items := make([]string, 0)

	if node == nil {
		return items
	}

	indent := strings.Repeat("  ", depth)
	icon := "📁"
	if !node.IsDir {
		icon = "📄"
	}

	items = append(items, fmt.Sprintf("%s%s %s", indent, icon, node.Name))

	// Add children if directory
	if node.IsDir {
		for _, child := range node.Children {
			items = append(items, p.flattenTree(child, depth+1)...)
		}
	}

	return items
}

// GetSelected returns the currently selected item
func (p *FileTreePane) GetSelected() int {
	return p.selected
}
