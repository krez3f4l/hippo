package service

import (
	"fmt"
)

type ValidationError struct {
	Field   string
	Message string
}

func NewValidationError(field, msg string) error {
	return &ValidationError{
		Field:   field,
		Message: msg,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

type NotFoundError struct {
	Entity string
	ID     interface{}
	Cause  error
}

func NewNotFoundError(entity string, id interface{}, err error) error {
	return &NotFoundError{
		Entity: entity,
		ID:     id,
		Cause:  err,
	}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%v", e.Cause)
}

type ErrDuplicateEmail struct {
	Cause error
}

func NewErrDuplicateEmail(cause error) error {
	return &ErrDuplicateEmail{cause}
}

func (e *ErrDuplicateEmail) Error() string {
	return fmt.Sprintf("email already exists: %s", e.Cause)
}

type ErrInvalidCredential struct {
	Cause error
}

func NewErrInvalidCredential(err error) error {
	return &ErrInvalidCredential{Cause: err}
}

func (e *ErrInvalidCredential) Error() string {
	return fmt.Sprintf("invalid credential")
}

type ErrRefreshTokenExpired struct{}

func NewErrRefreshTokenExpired() error {
	return &ErrRefreshTokenExpired{}
}

func (e *ErrRefreshTokenExpired) Error() string {
	return fmt.Sprintf("refresh token expired")
}
