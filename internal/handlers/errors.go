package handlers

import (
	"time"

	"github.com/google/uuid"
)

// StandardErrorResponse represents the new standardized error format
type StandardErrorResponse struct {
	Error ErrorDetail `json:"error"`
	Meta  ErrorMeta   `json:"meta"`
}

// ErrorDetail contains the error information
type ErrorDetail struct {
	Code    string                 `json:"code" example:"INVALID_INITIALS"`
	Message string                 `json:"message" example:"Player initials must be exactly 3 characters"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorMeta contains request metadata
type ErrorMeta struct {
	RequestID string `json:"request_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Timestamp string `json:"timestamp" example:"2025-07-16T15:30:00.000Z"`
}

// Error codes for consistent API responses
const (
	ErrorCodeInvalidInitials    = "INVALID_INITIALS"
	ErrorCodeInvalidScore       = "INVALID_SCORE"
	ErrorCodeInvalidGameID      = "INVALID_GAME_ID"
	ErrorCodeGameNotFound       = "GAME_NOT_FOUND"
	ErrorCodePlayerNotFound     = "PLAYER_NOT_FOUND"
	ErrorCodeScoreHistoryEmpty  = "SCORE_HISTORY_EMPTY"
	ErrorCodeValidationFailed   = "VALIDATION_FAILED"
	ErrorCodeAuthenticationRequired = "AUTHENTICATION_REQUIRED"
	ErrorCodeInvalidAPIKey      = "INVALID_API_KEY"
	ErrorCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
	ErrorCodeInvalidRequest     = "INVALID_REQUEST"
)

// NewStandardErrorResponse creates a standardized error response
func NewStandardErrorResponse(code, message string, details ...map[string]interface{}) *StandardErrorResponse {
	errorDetails := make(map[string]interface{})
	if len(details) > 0 && details[0] != nil {
		errorDetails = details[0]
	}

	return &StandardErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: errorDetails,
		},
		Meta: ErrorMeta{
			RequestID: uuid.New().String(),
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		},
	}
}

// NewValidationErrorResponse creates a validation error with field details
func NewValidationErrorResponse(field, value, constraint string) *StandardErrorResponse {
	return NewStandardErrorResponse(
		ErrorCodeValidationFailed,
		"Validation failed",
		map[string]interface{}{
			"field":      field,
			"value":      value,
			"constraint": constraint,
		},
	)
}
