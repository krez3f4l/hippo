package domain

import (
	"time"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"registered_at"`
}

type SignUpInfo struct {
	Name     string `json:"name" validate:"required,gte=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (i SignUpInfo) Validate() error {
	return validate.Struct(i)
}

type SignInInfo struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (i SignInInfo) Validate() error {
	return validate.Struct(i)
}
