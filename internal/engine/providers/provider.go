package providers

import (
	"context"
)

// ModelProvider defines the interface for LLM providers
type ModelProvider interface {
	// Name returns the provider name (e.g., "anthropic", "openai")
	Name() string

	// Request sends a prompt and returns the complete response
	Request(ctx context.Context, prompt string, opts RequestOptions) (string, error)

	// Stream sends a prompt and returns a channel of streaming tokens
	Stream(ctx context.Context, prompt string, opts RequestOptions) (<-chan string, <-chan error)

	// ListModels returns available models for this provider
	ListModels(ctx context.Context) ([]string, error)
}

// RequestOptions contains parameters for LLM requests
type RequestOptions struct {
	Model        string  // Model identifier
	Temperature  float64 // Sampling temperature (0.0 - 1.0)
	MaxTokens    int     // Maximum tokens to generate
	SystemPrompt string  // System prompt/instructions
}

// DefaultRequestOptions returns sensible defaults
func DefaultRequestOptions() RequestOptions {
	return RequestOptions{
		Temperature: 0.7,
		MaxTokens:   4000,
	}
}
