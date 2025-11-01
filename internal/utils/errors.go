package utils

import (
	"strings"
)

// IsUniqueViolationErr - Detects PostgreSQL unique constraint violation errors
// Purpose: Identifies database unique constraint violations from error messages for proper error handling
// Advantages:
//   - Database-agnostic error detection (works with PostgreSQL error format)
//   - Case-insensitive matching for reliability
//   - Nil-safe operation (handles nil errors gracefully)
//   - Simple boolean return for easy conditional logic
//
// Weaknesses:
//   - String-based detection (fragile if PostgreSQL changes error messages)
//   - PostgreSQL-specific (won't work with MySQL, SQLite, etc.)
//   - No distinction between different unique constraints
//   - Could produce false positives if error message contains the phrase
//
// Example:
//
//	if IsUniqueViolationErr(err) {
//	    return ErrDuplicateEmail
//	}
//	return ErrDatabase
func IsUniqueViolationErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value violates unique constraint")
}
