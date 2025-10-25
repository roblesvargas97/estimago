package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DecodeJSON - Safely decodes JSON request body into provided struct
// Purpose: Parses HTTP request body JSON with security measures and validation
// Advantages:
//   - Memory protection with MaxBytesReader (prevents DoS attacks)
//   - Strict field validation with DisallowUnknownFields
//   - Comprehensive error handling for different JSON parsing failures
//   - Prevents multiple JSON objects in single request
//
// Weaknesses:
//   - Fixed 1MB limit (not configurable)
//   - Error messages could be more user-friendly
//   - No support for streaming large JSON payloads
func DecodeJSON(r *http.Request, v any) error {

	// Security measure: Limit request body size to prevent memory exhaustion
	// Advantages: Prevents DoS attacks via large payloads
	// Weaknesses: Fixed limit, might be too restrictive for some use cases
	const maxBytes = 1 << 20 // 1 MB
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)

	// JSON decoder with strict validation
	// Advantages: Catches unexpected fields, maintains data integrity
	// Weaknesses: Less flexible, might reject valid but extended JSON
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	// Comprehensive error handling for JSON decoding failures
	// Advantages: Specific error messages for different failure types
	// Weaknesses: Error messages might expose internal structure to clients
	if err := dec.Decode(v); err != nil {
		var ute *json.UnmarshalTypeError
		switch {
		case errors.Is(err, io.EOF):
			return errors.New("empty body")
		case errors.As(err, &ute):
			return errors.New("wrong type for field: " + ute.Field)
		default:
			return err
		}
	}
	// Security check: Prevent multiple JSON objects in single request
	// Advantages: Prevents confusion and potential security issues
	// Weaknesses: Doesn't support JSON streaming or arrays of objects
	if dec.More() {
		return errors.New("multiple JSON objects in body")
	}
	return nil
}

// WriteJSON - Encodes and sends JSON response with proper headers
// Purpose: Standardizes JSON response writing with correct content type and status
// Advantages:
//   - Consistent JSON responses across the application
//   - Proper UTF-8 charset specification
//   - Simple interface for any serializable data
//   - Automatic JSON encoding with error handling
//
// Weaknesses:
//   - Ignores encoding errors (silently fails)
//   - No compression support (gzip/deflate)
//   - No pretty-printing option for debugging
//   - Fixed content type (can't customize headers)
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // Note: Error intentionally ignored
}

// WriteErr - Standardized error response writer
// Purpose: Creates consistent error responses with structured format
// Advantages:
//   - Uniform error response format across the API
//   - Separates error code from human-readable message
//   - Reuses WriteJSON for consistency
//   - Simple interface for error handling
//
// Weaknesses:
//   - Fixed error structure (not RFC 7807 compliant)
//   - No support for additional error metadata
//   - No localization support for error messages
//   - Limited error context (no request ID, timestamp, etc.)
func WriteErr(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, map[string]any{"error": code, "message": msg})
}
