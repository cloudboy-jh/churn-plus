package passes

import (
	"github.com/cloudboy-jh/churn-plus/internal/engine"
)

// CreateRefactorPass creates the Structural Refactor pass (Pass 2)
func CreateRefactorPass(provider, model string) *engine.Pass {
	return &engine.Pass{
		Name:        "refactor",
		Description: "Deep analysis for architectural improvements and refactoring opportunities",
		Status:      engine.PassPending,
		Model:       model,
		Provider:    provider,
	}
}
