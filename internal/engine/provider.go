package engine

import (
	"github.com/cloudboy-jh/churn-plus/internal/engine/providers"
)

// Re-export provider types to avoid import cycles
type ModelProvider = providers.ModelProvider
type RequestOptions = providers.RequestOptions

// DefaultRequestOptions returns sensible defaults
func DefaultRequestOptions() RequestOptions {
	return providers.DefaultRequestOptions()
}
