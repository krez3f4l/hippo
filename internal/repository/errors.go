package repository

import "fmt"

type NotFoundError struct {
	Op     string
	Entity string
	ID     interface{}
}

func NewNotFoundError(op, entity string, id interface{}) error {
	return &NotFoundError{
		Op:     op,
		Entity: entity,
		ID:     id,
	}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: %s with ID %v not found", e.Op, e.Entity, e.ID)
}

type ErrEmptyUpdate struct {
	Op     string
	Entity string
}

func NewErrEmptyUpdate(op, entity string) error {
	return &ErrEmptyUpdate{
		Op:     op,
		Entity: entity,
	}
}

func (e *ErrEmptyUpdate) Error() string {
	return fmt.Sprintf("%s: no fields provided for update %s", e.Op, e.Entity)
}

type ErrDuplicateEmail struct {
	Cause error
}

func NewErrDuplicateEmail(cause error) error {
	return &ErrDuplicateEmail{cause}
}

func (e *ErrDuplicateEmail) Error() string {
	return fmt.Sprintf("duplicated email: %s", e.Cause)
}

type ErrInvalidCredential struct{}

func NewErrInvalidCredential() error {
	return &ErrInvalidCredential{}
}

func (e *ErrInvalidCredential) Error() string {
	return fmt.Sprintf("invalid credential")
}

type ErrTokenNotFound struct{}

func NewErrTokenNotFound() error {
	return &ErrTokenNotFound{}
}

func (e *ErrTokenNotFound) Error() string {
	return fmt.Sprintf("invalid credential")
}
