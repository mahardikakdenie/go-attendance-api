package utils

import "fmt"

type AppError struct {
	Code    int
	Message string
	Details interface{}
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string, details interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(403, message, nil)
}

func NewNotFoundError(message string) *AppError {
	return NewAppError(404, message, nil)
}

func NewValidationError(message string, details interface{}) *AppError {
	return NewAppError(422, message, details)
}

func NewInternalError(message string, err error) *AppError {
	msg := message
	if err != nil {
		msg = fmt.Sprintf("%s: %v", message, err)
	}
	return NewAppError(500, msg, nil)
}
