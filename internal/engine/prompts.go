package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// BuildPromptForFile creates an analysis prompt for a file
func BuildPromptForFile(file *FileInfo, ctx *ProjectContext, pass *Pass) (string, error) {
	// Read file content
	content, err := os.ReadFile(file.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Build context information
	contextInfo := fmt.Sprintf(`Project Context:
- Root: %s
- Languages: %s
- Frameworks: %s
- File: %s (Language: %s, %d lines)
`,
		ctx.RootPath,
		strings.Join(ctx.Languages, ", "),
		strings.Join(ctx.Frameworks, ", "),
		file.Path,
		file.Language,
		file.Lines,
	)

	// Build analysis instructions based on pass type
	instructions := GetAnalysisInstructions(pass.Name, file.Language)

	// Combine into full prompt
	prompt := fmt.Sprintf(`%s

%s

File Content:
`+"```"+`%s
%s
`+"```"+`

Analyze this file and identify issues. Return your findings as a JSON array with this structure:
[
  {
    "line_start": <number>,
    "line_end": <number>,
    "severity": "low|medium|high|critical",
    "kind": "unused-import|unreachable-code|security|performance|etc",
    "message": "Description of the issue",
    "code": "Optional suggested fix"
  }
]

If no issues are found, return an empty array: []
`,
		contextInfo,
		instructions,
		file.Language,
		string(content),
	)

	return prompt, nil
}

// GetSystemPromptForPass returns the system prompt for a pass
func GetSystemPromptForPass(pass *Pass) string {
	switch pass.Name {
	case "lint":
		return "You are an expert code analyzer focused on identifying structural issues, unused code, and basic quality problems. Be precise and actionable."
	case "refactor":
		return "You are an expert software architect analyzing code for structural improvements, design patterns, and refactoring opportunities. Focus on maintainability and best practices."
	case "local-refinement":
		return "You are a code reviewer refining previous analysis results. Focus on validating findings and ensuring recommendations are practical."
	case "summary":
		return "You are synthesizing analysis results to ensure consistency and provide an overall assessment. Identify any conflicting recommendations."
	default:
		return "You are an expert code analyzer. Identify issues and suggest improvements."
	}
}

// GetAnalysisInstructions returns language and pass-specific instructions
func GetAnalysisInstructions(passName, language string) string {
	var instructions strings.Builder

	instructions.WriteString(fmt.Sprintf("Pass: %s\n\n", passName))

	switch passName {
	case "lint":
		instructions.WriteString("Focus on:\n")
		instructions.WriteString("- Unused imports, variables, and functions\n")
		instructions.WriteString("- Unreachable code\n")
		instructions.WriteString("- Basic syntax and structural issues\n")
		instructions.WriteString("- Type errors (if applicable)\n")

	case "refactor":
		instructions.WriteString("Focus on:\n")
		instructions.WriteString("- Code duplication\n")
		instructions.WriteString("- Complex functions that should be split\n")
		instructions.WriteString("- Opportunities for abstraction\n")
		instructions.WriteString("- Design pattern improvements\n")
		instructions.WriteString("- Performance optimizations\n")

	case "local-refinement":
		instructions.WriteString("Focus on:\n")
		instructions.WriteString("- Validating previous findings\n")
		instructions.WriteString("- Ensuring recommendations are practical\n")
		instructions.WriteString("- Identifying false positives\n")

	case "summary":
		instructions.WriteString("Focus on:\n")
		instructions.WriteString("- Overall code quality assessment\n")
		instructions.WriteString("- Consistency across findings\n")
		instructions.WriteString("- Priority ordering of issues\n")
	}

	// Add language-specific guidance
	instructions.WriteString(fmt.Sprintf("\nLanguage-specific considerations for %s:\n", language))

	switch language {
	case "typescript", "javascript":
		instructions.WriteString("- React hooks must be called in the same order every render\n")
		instructions.WriteString("- Async/await patterns and promise handling\n")
		instructions.WriteString("- Type safety (TypeScript)\n")
		instructions.WriteString("- Modern ES6+ patterns\n")

	case "python":
		instructions.WriteString("- Type hints and PEP 8 compliance\n")
		instructions.WriteString("- Pythonic idioms (list comprehensions, generators)\n")
		instructions.WriteString("- Async/await patterns\n")
		instructions.WriteString("- Exception handling\n")

	case "go":
		instructions.WriteString("- Error handling (check all errors)\n")
		instructions.WriteString("- Goroutine and channel patterns\n")
		instructions.WriteString("- Go idioms and conventions\n")
		instructions.WriteString("- Use of standard library\n")

	case "rust":
		instructions.WriteString("- Ownership and borrowing\n")
		instructions.WriteString("- Error handling (Result/Option)\n")
		instructions.WriteString("- Memory safety\n")
		instructions.WriteString("- Lifetime annotations\n")
	}

	return instructions.String()
}

// ParseFindingsFromResponse extracts findings from LLM response
func ParseFindingsFromResponse(filePath, response string) []*Finding {
	findings := make([]*Finding, 0)

	// Try to extract JSON array from response
	// LLMs sometimes wrap JSON in markdown code blocks
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return findings
	}

	var rawFindings []struct {
		LineStart int    `json:"line_start"`
		LineEnd   int    `json:"line_end"`
		Severity  string `json:"severity"`
		Kind      string `json:"kind"`
		Message   string `json:"message"`
		Code      string `json:"code"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawFindings); err != nil {
		// Failed to parse, return empty
		return findings
	}

	// Convert to Finding structs
	for _, rf := range rawFindings {
		severity := SeverityMedium
		switch strings.ToLower(rf.Severity) {
		case "low":
			severity = SeverityLow
		case "medium":
			severity = SeverityMedium
		case "high":
			severity = SeverityHigh
		case "critical":
			severity = SeverityCritical
		}

		findings = append(findings, &Finding{
			File:      filePath,
			LineStart: rf.LineStart,
			LineEnd:   rf.LineEnd,
			Severity:  severity,
			Kind:      rf.Kind,
			Message:   rf.Message,
			Code:      rf.Code,
		})
	}

	return findings
}

// extractJSON attempts to extract JSON from text (handles markdown code blocks)
func extractJSON(text string) string {
	// Try to find JSON array in markdown code block
	if strings.Contains(text, "```json") {
		start := strings.Index(text, "```json")
		if start != -1 {
			start += 7 // Skip "```json"
			end := strings.Index(text[start:], "```")
			if end != -1 {
				return strings.TrimSpace(text[start : start+end])
			}
		}
	}

	// Try to find JSON array directly
	if strings.Contains(text, "[") {
		start := strings.Index(text, "[")
		end := strings.LastIndex(text, "]")
		if start != -1 && end != -1 && end > start {
			return strings.TrimSpace(text[start : end+1])
		}
	}

	return ""
}
