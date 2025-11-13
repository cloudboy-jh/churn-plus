package languages

// PythonRules returns Python specific analysis rules
func PythonRules() []string {
	return []string{
		"Use type hints for function parameters and return values",
		"Follow PEP 8 style guidelines",
		"Use Pythonic idioms (list comprehensions, generators, context managers)",
		"Proper exception handling with specific exception types",
		"Avoid bare except clauses",
		"Use async/await for asynchronous operations",
		"Check for unused imports and variables",
		"Use f-strings for string formatting (Python 3.6+)",
		"Avoid mutable default arguments",
		"Use dataclasses or NamedTuple for data structures",
	}
}
