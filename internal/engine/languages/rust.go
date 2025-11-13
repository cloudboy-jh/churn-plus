package languages

// RustRules returns Rust specific analysis rules
func RustRules() []string {
	return []string{
		"Follow ownership and borrowing rules carefully",
		"Use Result<T, E> for error handling instead of panicking",
		"Leverage Option<T> for nullable values",
		"Avoid unnecessary cloning - use references where possible",
		"Use lifetime annotations correctly",
		"Follow Rust naming conventions (snake_case for functions/variables)",
		"Use pattern matching instead of if/else chains",
		"Implement proper Drop for resource cleanup",
		"Use iterators instead of index-based loops",
		"Check for memory safety issues",
		"Use cargo fmt and cargo clippy recommendations",
		"Avoid unsafe blocks unless absolutely necessary",
	}
}
