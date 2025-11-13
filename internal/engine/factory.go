package engine

import (
	"context"
	"fmt"

	"github.com/cloudboy-jh/churn-plus/internal/config"
	"github.com/cloudboy-jh/churn-plus/internal/engine/providers"
)

// Factory creates and configures engine components
type Factory struct {
	cfg *config.Config
}

// NewFactory creates a new engine factory
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{cfg: cfg}
}

// CreateProvider creates a model provider based on configuration
func (f *Factory) CreateProvider() (ModelProvider, error) {
	modelSelection := f.cfg.GetModelSelection()

	switch modelSelection.Provider {
	case "anthropic":
		apiKey := f.cfg.GetAPIKey("anthropic")
		if apiKey == "" {
			return nil, fmt.Errorf("anthropic API key not configured")
		}
		return providers.NewAnthropicProvider(apiKey), nil

	case "openai":
		apiKey := f.cfg.GetAPIKey("openai")
		if apiKey == "" {
			return nil, fmt.Errorf("openai API key not configured")
		}
		return providers.NewOpenAIProvider(apiKey), nil

	case "google":
		apiKey := f.cfg.GetAPIKey("google")
		if apiKey == "" {
			return nil, fmt.Errorf("google API key not configured")
		}
		return providers.NewGoogleProvider(apiKey), nil

	case "ollama":
		return providers.NewOllamaProvider(""), nil

	default:
		return nil, fmt.Errorf("unknown provider: %s", modelSelection.Provider)
	}
}

// CreateDefaultPipeline creates a pipeline with default or configured passes
func (f *Factory) CreateDefaultPipeline(provider ModelProvider) (*PipelineOrchestrator, error) {
	orchestrator := NewPipelineOrchestrator(provider)

	// Check if pipeline is configured in project config
	if f.cfg.Project.Pipeline != nil && len(f.cfg.Project.Pipeline.Passes) > 0 {
		// Use configured pipeline
		for _, passConfig := range f.cfg.Project.Pipeline.Passes {
			if passConfig.Enabled {
				orchestrator.AddPass(&Pass{
					Name:        passConfig.Name,
					Description: passConfig.Description,
					Status:      PassPending,
					Model:       passConfig.Model,
					Provider:    passConfig.Provider,
				})
			}
		}
		return orchestrator, nil
	}

	// No pipeline configured, use defaults
	modelSelection := f.cfg.GetModelSelection()

	// Add default passes
	// Pass 1: Lint (use fast model)
	lintModel := "claude-3-5-haiku-20241022"
	if modelSelection.Provider == "openai" {
		lintModel = "gpt-3.5-turbo"
	} else if modelSelection.Provider == "ollama" {
		lintModel = f.getFirstOllamaModel(provider)
	}
	orchestrator.AddPass(&Pass{
		Name:        "lint",
		Description: "Quick structural checks for unused code and basic issues",
		Status:      PassPending,
		Model:       lintModel,
		Provider:    modelSelection.Provider,
	})

	// Pass 2: Refactor (use main model)
	orchestrator.AddPass(&Pass{
		Name:        "refactor",
		Description: "Deep analysis for architectural improvements and refactoring opportunities",
		Status:      PassPending,
		Model:       modelSelection.Model,
		Provider:    modelSelection.Provider,
	})

	// Pass 3: Local refinement (optional, only if Ollama available)
	if modelSelection.Provider == "ollama" {
		orchestrator.AddPass(&Pass{
			Name:        "local-refinement",
			Description: "Optional local model refinement for privacy-focused validation",
			Status:      PassPending,
			Model:       lintModel,
			Provider:    "ollama",
		})
	}

	// Pass 4: Summary
	orchestrator.AddPass(&Pass{
		Name:        "summary",
		Description: "Ensures coherence across findings and provides overall assessment",
		Status:      PassPending,
		Model:       modelSelection.Model,
		Provider:    modelSelection.Provider,
	})

	return orchestrator, nil
}

// getFirstOllamaModel gets the first available Ollama model
func (f *Factory) getFirstOllamaModel(provider ModelProvider) string {
	ctx := context.Background()
	models, err := provider.ListModels(ctx)
	if err != nil || len(models) == 0 {
		return "llama2" // Fallback
	}
	return models[0]
}

// ScanProject scans a project directory
func (f *Factory) ScanProject(projectRoot string) ([]*FileInfo, *FileNode, error) {
	scanner := NewScanner(projectRoot, f.cfg.Project.IgnorePatterns)
	files, err := scanner.Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan project: %w", err)
	}

	tree := BuildFileTree(files, projectRoot)

	return files, tree, nil
}

// BuildContext builds project context from scanned files
func (f *Factory) BuildContext(projectRoot string, files []*FileInfo) *ProjectContext {
	builder := NewContextBuilder(projectRoot)
	return builder.Build(files)
}
