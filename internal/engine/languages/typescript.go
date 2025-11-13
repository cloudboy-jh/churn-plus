package languages

// TypeScriptRules returns TypeScript/JavaScript specific analysis rules
func TypeScriptRules() []string {
	return []string{
		"React hooks must be called in the same order every render",
		"Async/await patterns should handle errors properly",
		"Type safety: avoid 'any' types where possible",
		"Use modern ES6+ syntax (const/let, arrow functions, destructuring)",
		"Check for unused imports and variables",
		"Avoid callback hell - use async/await or promises",
		"Ensure proper error boundaries in React components",
	}
}

// JavaScriptRules returns JavaScript specific analysis rules
func JavaScriptRules() []string {
	return TypeScriptRules() // Similar rules for JS
}
