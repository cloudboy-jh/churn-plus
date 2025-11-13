package engine

import (
	"bufio"
	"fmt"
	"strings"
)

// DiffEngine generates diffs between original and suggested code
type DiffEngine struct{}

// NewDiffEngine creates a new diff engine
func NewDiffEngine() *DiffEngine {
	return &DiffEngine{}
}

// Diff represents a unified diff
type Diff struct {
	FilePath string
	Hunks    []*DiffHunk
}

// DiffHunk represents a single hunk in a diff
type DiffHunk struct {
	OriginalStart int
	OriginalLines int
	ModifiedStart int
	ModifiedLines int
	Lines         []*DiffLine
}

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    DiffLineType // Added, Removed, Context
	Content string
	LineNum int // Original line number
}

// DiffLineType represents the type of diff line
type DiffLineType string

const (
	DiffLineAdded   DiffLineType = "added"
	DiffLineRemoved DiffLineType = "removed"
	DiffLineContext DiffLineType = "context"
)

// Generate creates a unified diff between original and modified content
func (de *DiffEngine) Generate(filePath, original, modified string) (*Diff, error) {
	originalLines := splitLines(original)
	modifiedLines := splitLines(modified)

	// Simple line-by-line diff (can be enhanced with proper diff algorithm like Myers)
	hunks := de.generateHunks(originalLines, modifiedLines)

	return &Diff{
		FilePath: filePath,
		Hunks:    hunks,
	}, nil
}

// generateHunks creates diff hunks from original and modified lines
func (de *DiffEngine) generateHunks(original, modified []string) []*DiffHunk {
	// This is a simplified diff implementation
	// For production, consider using a proper diff library like go-diff

	if len(original) == 0 && len(modified) == 0 {
		return []*DiffHunk{}
	}

	// Create a single hunk for simplicity
	hunk := &DiffHunk{
		OriginalStart: 1,
		OriginalLines: len(original),
		ModifiedStart: 1,
		ModifiedLines: len(modified),
		Lines:         make([]*DiffLine, 0),
	}

	// Simple implementation: mark all original lines as removed, all new as added
	maxLen := len(original)
	if len(modified) > maxLen {
		maxLen = len(modified)
	}

	for i := 0; i < maxLen; i++ {
		if i < len(original) && i < len(modified) {
			if original[i] == modified[i] {
				// Context line
				hunk.Lines = append(hunk.Lines, &DiffLine{
					Type:    DiffLineContext,
					Content: original[i],
					LineNum: i + 1,
				})
			} else {
				// Line changed
				hunk.Lines = append(hunk.Lines, &DiffLine{
					Type:    DiffLineRemoved,
					Content: original[i],
					LineNum: i + 1,
				})
				hunk.Lines = append(hunk.Lines, &DiffLine{
					Type:    DiffLineAdded,
					Content: modified[i],
					LineNum: i + 1,
				})
			}
		} else if i < len(original) {
			// Line removed
			hunk.Lines = append(hunk.Lines, &DiffLine{
				Type:    DiffLineRemoved,
				Content: original[i],
				LineNum: i + 1,
			})
		} else {
			// Line added
			hunk.Lines = append(hunk.Lines, &DiffLine{
				Type:    DiffLineAdded,
				Content: modified[i],
				LineNum: len(original) + (i - len(original)) + 1,
			})
		}
	}

	return []*DiffHunk{hunk}
}

// FormatUnified formats a diff in unified diff format
func (d *Diff) FormatUnified() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("--- a/%s\n", d.FilePath))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", d.FilePath))

	for _, hunk := range d.Hunks {
		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			hunk.OriginalStart, hunk.OriginalLines,
			hunk.ModifiedStart, hunk.ModifiedLines))

		for _, line := range hunk.Lines {
			prefix := " "
			switch line.Type {
			case DiffLineAdded:
				prefix = "+"
			case DiffLineRemoved:
				prefix = "-"
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", prefix, line.Content))
		}
	}

	return sb.String()
}

// GetChangedLines returns a map of line numbers that were changed
func (d *Diff) GetChangedLines() map[int]bool {
	changed := make(map[int]bool)

	for _, hunk := range d.Hunks {
		for _, line := range hunk.Lines {
			if line.Type == DiffLineAdded || line.Type == DiffLineRemoved {
				changed[line.LineNum] = true
			}
		}
	}

	return changed
}

// GetChangeCount returns the number of additions and deletions
func (d *Diff) GetChangeCount() (additions int, deletions int) {
	for _, hunk := range d.Hunks {
		for _, line := range hunk.Lines {
			switch line.Type {
			case DiffLineAdded:
				additions++
			case DiffLineRemoved:
				deletions++
			}
		}
	}
	return
}

// splitLines splits content into lines
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

// ApplyFindingSuggestion applies a finding's suggestion to generate a diff
func ApplyFindingSuggestion(finding *Finding, originalContent string) (*Diff, error) {
	if finding.Code == "" {
		// No code suggestion provided
		return nil, fmt.Errorf("finding has no code suggestion")
	}

	// For now, create a simple diff showing the suggestion
	// In a real implementation, this would intelligently apply the change
	lines := splitLines(originalContent)

	// Replace lines in the specified range
	modifiedLines := make([]string, 0, len(lines))
	modifiedLines = append(modifiedLines, lines[:finding.LineStart-1]...)
	modifiedLines = append(modifiedLines, finding.Code)
	if finding.LineEnd < len(lines) {
		modifiedLines = append(modifiedLines, lines[finding.LineEnd:]...)
	}

	engine := NewDiffEngine()
	return engine.Generate(finding.File, originalContent, strings.Join(modifiedLines, "\n"))
}
