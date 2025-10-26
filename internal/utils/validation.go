package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("not_empty", notEmpty)
	v.RegisterValidation("vector_dimension", vectorDimension)
	
	return &Validator{validator: v}
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validator.Var(field, tag)
}

func (v *Validator) GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			errors[field] = getErrorMessage(e)
		}
	}
	
	return errors
}

func notEmpty(fl validator.FieldLevel) bool {
	field := fl.Field()
	
	switch field.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return field.Len() > 0
	case reflect.String:
		return strings.TrimSpace(field.String()) != ""
	default:
		return true
	}
}

func vectorDimension(fl validator.FieldLevel) bool {
	field := fl.Field()
	
	if field.Kind() != reflect.Slice {
		return false
	}
	
	// Check if it's a slice of float64
	if field.Type().Elem().Kind() != reflect.Float64 {
		return false
	}
	
	length := field.Len()
	
	// Vector dimension should be between 1 and 10000
	return length >= 1 && length <= 10000
}

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "not_empty":
		return fmt.Sprintf("%s cannot be empty", fe.Field())
	case "vector_dimension":
		return fmt.Sprintf("%s must be a valid vector with dimension between 1 and 10000", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

func ValidateStruct(s interface{}) error {
	v := NewValidator()
	return v.Validate(s)
}

func ValidateStructWithDetails(s interface{}) map[string]string {
	v := NewValidator()
	err := v.Validate(s)
	if err == nil {
		return nil
	}
	return v.GetValidationErrors(err)
}
