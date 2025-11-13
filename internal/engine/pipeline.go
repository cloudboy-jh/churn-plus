package engine

import (
	"context"
	"fmt"
	"time"
)

// PipelineOrchestrator manages the execution of analysis passes
type PipelineOrchestrator struct {
	pipeline *Pipeline
	provider ModelProvider
	events   chan PipelineEvent
}

// NewPipelineOrchestrator creates a new pipeline orchestrator
func NewPipelineOrchestrator(provider ModelProvider) *PipelineOrchestrator {
	return &PipelineOrchestrator{
		pipeline: &Pipeline{
			Passes:    make([]*Pass, 0),
			Findings:  make([]*Finding, 0),
			StartTime: time.Now(),
		},
		provider: provider,
		events:   make(chan PipelineEvent, 100),
	}
}

// AddPass adds a pass to the pipeline
func (po *PipelineOrchestrator) AddPass(pass *Pass) {
	po.pipeline.Passes = append(po.pipeline.Passes, pass)
}

// SetContext sets the project context
func (po *PipelineOrchestrator) SetContext(ctx *ProjectContext) {
	po.pipeline.Context = ctx
}

// Events returns the event channel for subscribing to pipeline updates
func (po *PipelineOrchestrator) Events() <-chan PipelineEvent {
	return po.events
}

// Execute runs the pipeline
func (po *PipelineOrchestrator) Execute(ctx context.Context, files []*FileInfo) error {
	defer close(po.events)

	for _, pass := range po.pipeline.Passes {
		if err := po.executePass(ctx, pass, files); err != nil {
			pass.Status = PassFailed
			pass.Error = err.Error()
			pass.EndTime = time.Now()

			po.events <- PipelineEvent{
				Type:  EventPassFailed,
				Pass:  pass,
				Error: err,
			}

			return fmt.Errorf("pass %s failed: %w", pass.Name, err)
		}
	}

	po.pipeline.EndTime = time.Now()
	return nil
}

// executePass runs a single pass
func (po *PipelineOrchestrator) executePass(ctx context.Context, pass *Pass, files []*FileInfo) error {
	pass.Status = PassRunning
	pass.StartTime = time.Now()

	po.events <- PipelineEvent{
		Type: EventPassStarted,
		Pass: pass,
	}

	// Execute the pass based on its type
	findings, err := po.runPassAnalysis(ctx, pass, files)
	if err != nil {
		return err
	}

	// Add findings to pipeline
	for _, finding := range findings {
		finding.Pass = pass.Name
		po.pipeline.Findings = append(po.pipeline.Findings, finding)

		po.events <- PipelineEvent{
			Type:    EventFindingAdded,
			Finding: finding,
		}
	}

	pass.Status = PassCompleted
	pass.EndTime = time.Now()

	po.events <- PipelineEvent{
		Type: EventPassCompleted,
		Pass: pass,
	}

	return nil
}

// runPassAnalysis performs the actual analysis for a pass
func (po *PipelineOrchestrator) runPassAnalysis(ctx context.Context, pass *Pass, files []*FileInfo) ([]*Finding, error) {
	findings := make([]*Finding, 0)

	// For each file, send to LLM for analysis
	for _, file := range files {
		// Send progress event
		po.events <- PipelineEvent{
			Type:    EventPassProgress,
			Pass:    pass,
			Message: fmt.Sprintf("Analyzing %s", file.Path),
		}

		// Build prompt for this file
		prompt, err := BuildPromptForFile(file, po.pipeline.Context, pass)
		if err != nil {
			continue // Skip files we can't build prompts for
		}

		// Request analysis from LLM
		opts := DefaultRequestOptions()
		opts.Model = pass.Model
		opts.SystemPrompt = GetSystemPromptForPass(pass)

		response, err := po.provider.Request(ctx, prompt, opts)
		if err != nil {
			// Log error but continue with other files
			continue
		}

		// Parse findings from response
		fileFindings := ParseFindingsFromResponse(file.Path, response)
		findings = append(findings, fileFindings...)
	}

	return findings, nil
}

// GetPipeline returns the pipeline
func (po *PipelineOrchestrator) GetPipeline() *Pipeline {
	return po.pipeline
}

// GetFindings returns all findings
func (po *PipelineOrchestrator) GetFindings() []*Finding {
	return po.pipeline.Findings
}
