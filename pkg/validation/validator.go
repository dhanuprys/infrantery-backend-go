package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationEngine handles struct validation and error formatting
type ValidationEngine struct {
	validate *validator.Validate
}

// NewValidationEngine creates a new instance of ValidationEngine
func NewValidationEngine() *ValidationEngine {
	v := validator.New()

	// Register custom tag name function to use "json" tag for field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &ValidationEngine{
		validate: v,
	}
}

// ValidateStruct validates a struct and returns formatted error fields if validation fails
func (ve *ValidationEngine) ValidateStruct(s interface{}) *[]map[string]string {
	err := ve.validate.Struct(s)
	if err == nil {
		return nil
	}

	var veErrors validator.ValidationErrors
	if errors.As(err, &veErrors) {
		out := make([]map[string]string, len(veErrors))
		for i, fe := range veErrors {
			out[i] = map[string]string{
				// Field returns the value from the registered TagNameFunc (json tag)
				fe.Field(): msgForTag(fe),
			}
		}
		return &out
	}

	return nil
}

// msgForTag returns a friendly error message
func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Minimum length is %s", fe.Param())
	case "max":
		return fmt.Sprintf("Maximum length is %s", fe.Param())
	case "uuid":
		return "Invalid UUID format"
	case "alpha":
		return "Must contain only letters"
	case "alphanum":
		return "Must contain only letters and numbers"
	case "numeric":
		return "Must be valid numeric value"
	case "len":
		return fmt.Sprintf("Length must be exactly %s", fe.Param())
	}
	return fe.Error() // Default error message
}
