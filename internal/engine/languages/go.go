package languages

// GoRules returns Go specific analysis rules
func GoRules() []string {
	return []string{
		"Always check and handle errors - never ignore error returns",
		"Use goroutines and channels properly - avoid race conditions",
		"Follow Go idioms and conventions (gofmt, golint)",
		"Use defer for cleanup operations",
		"Avoid naked returns in long functions",
		"Use context.Context for cancellation and timeouts",
		"Close resources properly (files, connections, etc.)",
		"Use Go's standard library effectively",
		"Avoid pointer to interface",
		"Use meaningful variable names (not single letters except for short scopes)",
		"Check slice bounds before accessing",
		"Use sync.WaitGroup or channels for goroutine synchronization",
	}
}
