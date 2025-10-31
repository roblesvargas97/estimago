package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DecodeJSON - Safely decodes JSON request body into provided struct
// Purpose:
//
//	Parses an HTTP request body as JSON with security measures and validation.
//
// Advantages:
//   - Memory protection with MaxBytesReader (prevents DoS attacks)
//   - Strict field validation with DisallowUnknownFields
//   - Comprehensive error handling for different JSON parsing failures
//   - Prevents multiple JSON objects in a single request
//
// Weaknesses:
//   - Fixed 1MB limit (not configurable)
//   - Error messages could be more user-friendly
//   - No support for streaming large JSON payloads
//
// Parameters:
//   - w: http.ResponseWriter (used by MaxBytesReader to enforce body size limits)
//   - r: *http.Request containing the JSON body
//   - v: pointer to a struct or map where decoded data will be stored
//
// Security Notes:
//   - The MaxBytesReader ensures payloads larger than 1MB are rejected before decoding.
//   - DisallowUnknownFields enforces strict schema matching, avoiding accidental data injection.
//
// Example:
//
//	var input CreateUserInput
//	if err := httputil.DecodeJSON(w, r, &input); err != nil {
//	    httputil.WriteErr(w, http.StatusBadRequest, "invalid_json", err.Error())
//	    return
//	}
//
// Behavior:
//   - Returns an error if:
//   - The body is empty
//   - The JSON contains fields not defined in the struct
//   - The JSON contains multiple top-level objects
//   - A field type does not match the expected Go type
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) error {

	const maxBytes = 1 << 20 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

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
