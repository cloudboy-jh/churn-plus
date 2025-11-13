package passes

import (
	"github.com/cloudboy-jh/churn-plus/internal/engine"
)

// CreateSummaryPass creates the Consistency & Summary pass (Pass 4)
func CreateSummaryPass(provider, model string) *engine.Pass {
	return &engine.Pass{
		Name:        "summary",
		Description: "Ensures coherence across findings and provides overall assessment",
		Status:      engine.PassPending,
		Model:       model,
		Provider:    provider,
	}
}
