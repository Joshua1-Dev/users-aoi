package utils

import (
	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidationErrors(err error) []ValidationError {
	var errs []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		e := ValidationError{
			Field:   err.Field(),
			Message: err.Error(),
		}
		errs = append(errs, e)
	}
	return errs
}
