package shared

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("slug", validateSlug)
}

func validateSlug(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

type IDParam struct {
	Value string `validate:"required,max=50,slug"`
}

func ValidateID(name, value string) error {
	param := IDParam{Value: value}
	if err := validate.Struct(param); err != nil {
		return formatValidationError(name, err)
	}
	return nil
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func formatValidationError(field string, err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				return fmt.Errorf("%s is required", field)
			case "max":
				return fmt.Errorf("%s cannot exceed %s characters", field, e.Param())
			case "min":
				return fmt.Errorf("%s must be at least %s", field, e.Param())
			case "slug":
				return fmt.Errorf("%s contains invalid characters (allowed: a-z, A-Z, 0-9, -, _)", field)
			case "oneof":
				return fmt.Errorf("%s must be one of: %s", field, e.Param())
			default:
				return fmt.Errorf("%s is invalid", field)
			}
		}
	}
	return err
}

func FormatValidationErrors(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errors = append(errors, fmt.Sprintf("%s is required", field))
			case "max":
				errors = append(errors, fmt.Sprintf("%s cannot exceed %s", field, e.Param()))
			case "min":
				errors = append(errors, fmt.Sprintf("%s must be at least %s", field, e.Param()))
			case "oneof":
				errors = append(errors, fmt.Sprintf("%s must be one of: %s", field, e.Param()))
			default:
				errors = append(errors, fmt.Sprintf("%s is invalid", field))
			}
		}
	}
	return errors
}
