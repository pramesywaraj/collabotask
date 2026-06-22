package errors

const (
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeInternal           = "INTERNAL_ERROR"
)

type AppError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(statusCode int, code, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}
