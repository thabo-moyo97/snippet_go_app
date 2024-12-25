package validator

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Form represents a form with validation capabilities
type Form struct {
	Errors map[string]string
}

// Valid returns true if there are no validation errors
func (f *Form) Valid() bool {
	return f.Errors == nil || len(f.Errors) == 0
}

// AddError adds an error message for a given field
func (f *Form) AddError(field, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	if _, exists := f.Errors[field]; !exists {
		f.Errors[field] = message
	}
}

// AddNonFieldError adds a general error message not tied to a specific field
func (f *Form) AddNonFieldError(message string) {
	f.AddError("generic", message)
}

type Validator struct {
	validate *validator.Validate
	db       *sql.DB
}

func New(db *sql.DB) *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterValidation("wordcount", func(fl validator.FieldLevel) bool {
		count := len(strings.Fields(fl.Field().String()))
		min, err := strconv.Atoi(fl.Param())
		if err != nil {
			return false
		}
		return count >= min
	})

	v.RegisterValidation("exists", func(fl validator.FieldLevel) bool {
		params := strings.Split(fl.Param(), ":")
		if len(params) != 2 {
			return false
		}
		table, column := params[0], params[1]

		value := fl.Field().Interface()
		var exists bool
		query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", table, column)
		err := db.QueryRow(query, value).Scan(&exists)

		defer func() bool {
			if r := recover(); r != nil || err != nil {
				return !exists
			}

			return err == nil && exists
		}()

		return err == nil && exists
	})

	return &Validator{validate: v, db: db}
}

// Validate validates a struct and returns a Form with any validation errors
func (v *Validator) Validate(i interface{}) *Form {
	form := &Form{
		Errors: make(map[string]string),
	}

	err := v.validate.Struct(i)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := err.StructField()
			// Get the form tag name if it exists
			if t, ok := reflect.TypeOf(i).Elem().FieldByName(field); ok {
				if formTag := t.Tag.Get("form"); formTag != "" {
					field = formTag
				}
			}
			form.Errors[field] = formatError(err)
		}
	}

	return form
}

func formatError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return fmt.Sprintf("This field must be at least %s characters long", err.Param())
	case "max":
		return fmt.Sprintf("This field cannot be longer than %s characters", err.Param())
	case "email":
		return "This field must be a valid email address"
	case "oneof":
		return fmt.Sprintf("This field must be one of: %s", err.Param())
	case "wordcount":
		return fmt.Sprintf("This field must contain at least %s words", err.Param())
	default:
		return fmt.Sprintf("Validation failed on condition: %s", err.Tag())
	}
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
