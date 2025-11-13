package languages

// ReactRules returns React specific analysis rules
func ReactRules() []string {
	return []string{
		"React hooks must be called at the top level - never in loops, conditions, or nested functions",
		"Use useEffect cleanup functions to prevent memory leaks",
		"Memoize expensive calculations with useMemo",
		"Use useCallback for event handlers passed to optimized child components",
		"Avoid inline object/array creation in JSX props",
		"Use keys properly in lists - avoid using index as key",
		"Implement error boundaries for component error handling",
		"Use React.memo for expensive pure components",
		"Avoid prop drilling - consider Context API or state management",
		"Keep component logic separate from UI rendering",
		"Use functional components and hooks over class components",
		"Ensure all dependencies are listed in useEffect/useMemo/useCallback arrays",
	}
}
