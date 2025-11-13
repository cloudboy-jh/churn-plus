package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ContextBuilder builds project context from scanned files
type ContextBuilder struct {
	rootPath string
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(rootPath string) *ContextBuilder {
	return &ContextBuilder{
		rootPath: rootPath,
	}
}

// Build creates a ProjectContext from scanned files
func (cb *ContextBuilder) Build(files []*FileInfo) *ProjectContext {
	ctx := &ProjectContext{
		RootPath:     cb.rootPath,
		Languages:    cb.detectLanguages(files),
		Frameworks:   cb.detectFrameworks(),
		Tools:        cb.detectTools(),
		Dependencies: cb.extractDependencies(),
		FileCount:    len(files),
	}

	return ctx
}

// detectLanguages identifies all languages in the project
func (cb *ContextBuilder) detectLanguages(files []*FileInfo) []string {
	langMap := make(map[string]bool)

	for _, file := range files {
		if file.Language != "unknown" {
			langMap[file.Language] = true
		}
	}

	languages := make([]string, 0, len(langMap))
	for lang := range langMap {
		languages = append(languages, lang)
	}

	return languages
}

// detectFrameworks identifies frameworks used in the project
func (cb *ContextBuilder) detectFrameworks() []string {
	frameworks := make([]string, 0)

	// Check package.json for JS/TS frameworks
	packageJSON := filepath.Join(cb.rootPath, "package.json")
	if data, err := os.ReadFile(packageJSON); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			deps := make(map[string]bool)

			if d, ok := pkg["dependencies"].(map[string]interface{}); ok {
				for k := range d {
					deps[k] = true
				}
			}
			if d, ok := pkg["devDependencies"].(map[string]interface{}); ok {
				for k := range d {
					deps[k] = true
				}
			}

			// Detect popular frameworks
			if deps["react"] {
				frameworks = append(frameworks, "React")
			}
			if deps["next"] {
				frameworks = append(frameworks, "Next.js")
			}
			if deps["vue"] {
				frameworks = append(frameworks, "Vue")
			}
			if deps["@angular/core"] {
				frameworks = append(frameworks, "Angular")
			}
			if deps["svelte"] {
				frameworks = append(frameworks, "Svelte")
			}
			if deps["express"] {
				frameworks = append(frameworks, "Express")
			}
			if deps["@nestjs/core"] {
				frameworks = append(frameworks, "NestJS")
			}
		}
	}

	// Check for Go frameworks
	goMod := filepath.Join(cb.rootPath, "go.mod")
	if data, err := os.ReadFile(goMod); err == nil {
		content := string(data)
		if strings.Contains(content, "github.com/gin-gonic/gin") {
			frameworks = append(frameworks, "Gin")
		}
		if strings.Contains(content, "github.com/gorilla/mux") {
			frameworks = append(frameworks, "Gorilla Mux")
		}
		if strings.Contains(content, "github.com/labstack/echo") {
			frameworks = append(frameworks, "Echo")
		}
	}

	// Check for Python frameworks
	requirementsTxt := filepath.Join(cb.rootPath, "requirements.txt")
	if data, err := os.ReadFile(requirementsTxt); err == nil {
		content := string(data)
		if strings.Contains(content, "django") {
			frameworks = append(frameworks, "Django")
		}
		if strings.Contains(content, "flask") {
			frameworks = append(frameworks, "Flask")
		}
		if strings.Contains(content, "fastapi") {
			frameworks = append(frameworks, "FastAPI")
		}
	}

	return frameworks
}

// detectTools identifies development tools used
func (cb *ContextBuilder) detectTools() []string {
	tools := make([]string, 0)

	// Check for common config files
	configs := map[string]string{
		"package.json":      "npm/yarn",
		"tsconfig.json":     "TypeScript",
		"Cargo.toml":        "Cargo",
		"go.mod":            "Go Modules",
		"requirements.txt":  "pip",
		"Pipfile":           "Pipenv",
		"poetry.lock":       "Poetry",
		".eslintrc":         "ESLint",
		".prettierrc":       "Prettier",
		"jest.config.js":    "Jest",
		"vitest.config.ts":  "Vitest",
		"webpack.config.js": "Webpack",
		"vite.config.ts":    "Vite",
	}

	for file, tool := range configs {
		if _, err := os.Stat(filepath.Join(cb.rootPath, file)); err == nil {
			tools = append(tools, tool)
		}
	}

	return tools
}

// extractDependencies reads key dependencies
func (cb *ContextBuilder) extractDependencies() map[string]string {
	deps := make(map[string]string)

	// Extract from package.json
	packageJSON := filepath.Join(cb.rootPath, "package.json")
	if data, err := os.ReadFile(packageJSON); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			if d, ok := pkg["dependencies"].(map[string]interface{}); ok {
				for k, v := range d {
					if ver, ok := v.(string); ok {
						deps[k] = ver
					}
				}
			}
		}
	}

	return deps
}
