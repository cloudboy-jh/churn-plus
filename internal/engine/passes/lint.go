package passes

import (
	"github.com/cloudboy-jh/churn-plus/internal/engine"
)

// CreateLintPass creates the Lint/Sanity pass (Pass 1)
func CreateLintPass(provider, model string) *engine.Pass {
	return &engine.Pass{
		Name:        "lint",
		Description: "Quick structural checks for unused code and basic issues",
		Status:      engine.PassPending,
		Model:       model,
		Provider:    provider,
	}
}
