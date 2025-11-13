package passes

import (
	"github.com/cloudboy-jh/churn-plus/internal/engine"
)

// CreateLocalRefinementPass creates the Local Refinement pass (Pass 3)
func CreateLocalRefinementPass(provider, model string) *engine.Pass {
	return &engine.Pass{
		Name:        "local-refinement",
		Description: "Optional local model refinement for privacy-focused validation",
		Status:      engine.PassPending,
		Model:       model,
		Provider:    provider,
	}
}
