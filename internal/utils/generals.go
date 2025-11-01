package utils

// DefaultIfEmpty - Returns default value if input string is empty
// Purpose: Provides fallback values for empty strings, common in configuration and validation
// Advantages:
//   - Simple and readable null-coalescing behavior
//   - Prevents empty string propagation in business logic
//   - Zero dependencies, pure Go implementation
//   - Consistent behavior across the application
// Weaknesses:
//   - Only handles empty strings, not whitespace-only strings
//   - No support for other "falsy" values (nil, zero values)
//   - Limited to string types only
// Example:
//   port := DefaultIfEmpty(os.Getenv("PORT"), "8080")
//   name := DefaultIfEmpty(user.Name, "Anonymous")
func DefaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
