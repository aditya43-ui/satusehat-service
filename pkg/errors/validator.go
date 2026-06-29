package errors

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Rule    string      `json:"rule"`
	Message string      `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	if len(ve) == 1 {
		return fmt.Sprintf("validation failed: %s", ve[0].Message)
	}

	return fmt.Sprintf("validation failed: %d errors", len(ve))
}

// ToMap converts validation errors to map
func (ve ValidationErrors) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for _, err := range ve {
		result[err.Field] = map[string]interface{}{
			"value":   err.Value,
			"rule":    err.Rule,
			"message": err.Message,
		}
	}
	return result
}

// Validator provides validation functionality
type Validator struct {
	rules map[string]ValidationRule
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name      string
	Validator func(interface{}) error
	Message   string
}

// NewValidator creates new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string]ValidationRule),
	}
}

// AddRule adds validation rule
func (v *Validator) AddRule(name string, validator func(interface{}) error, message string) {
	v.rules[name] = ValidationRule{
		Name:      name,
		Validator: validator,
		Message:   message,
	}
}

// Validate validates value against rules
func (v *Validator) Validate(value interface{}, rules []string) ValidationErrors {
	var errors ValidationErrors

	for _, ruleName := range rules {
		if rule, exists := v.rules[ruleName]; exists {
			if err := rule.Validator(value); err != nil {
				errors = append(errors, ValidationError{
					Field:   "",
					Value:   value,
					Rule:    ruleName,
					Message: rule.Message,
				})
			}
		}
	}

	return errors
}

// ValidateField validates a field with rules
func (v *Validator) ValidateField(field string, value interface{}, rules []string) ValidationErrors {
	var errors ValidationErrors

	for _, ruleName := range rules {
		if rule, exists := v.rules[ruleName]; exists {
			if err := rule.Validator(value); err != nil {
				errors = append(errors, ValidationError{
					Field:   field,
					Value:   value,
					Rule:    ruleName,
					Message: fmt.Sprintf(rule.Message, field),
				})
			}
		}
	}

	return errors
}

// ValidateStruct validates struct fields
func (v *Validator) ValidateStruct(obj interface{}) ValidationErrors {
	var errors ValidationErrors

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Get validation tags
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		fieldErrors := v.ValidateField(field.Name, fieldValue.Interface(), rules)
		errors = append(errors, fieldErrors...)
	}

	return errors
}

// Common validation rules
var (
	// Required rule
	Required = ValidationRule{
		Name: "required",
		Validator: func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("value is required")
			}
			return nil
		},
		Message: "%s is required",
	}

	// Email rule
	Email = ValidationRule{
		Name: "email",
		Validator: func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("value must be string")
			}

			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(str) {
				return fmt.Errorf("invalid email format")
			}
			return nil
		},
		Message: "%s must be a valid email",
	}

	// Min length rule
	MinLength = func(min int) ValidationRule {
		return ValidationRule{
			Name: fmt.Sprintf("min_%d", min),
			Validator: func(value interface{}) error {
				str, ok := value.(string)
				if !ok {
					return fmt.Errorf("value must be string")
				}

				if len(str) < min {
					return fmt.Errorf("value must be at least %d characters", min)
				}
				return nil
			},
			Message: fmt.Sprintf("%%s must be at least %d characters", min),
		}
	}

	// Max length rule
	MaxLength = func(max int) ValidationRule {
		return ValidationRule{
			Name: fmt.Sprintf("max_%d", max),
			Validator: func(value interface{}) error {
				str, ok := value.(string)
				if !ok {
					return fmt.Errorf("value must be string")
				}

				if len(str) > max {
					return fmt.Errorf("value must be at most %d characters", max)
				}
				return nil
			},
			Message: fmt.Sprintf("%%s must be at most %d characters", max),
		}
	}

	// Numeric rule
	Numeric = ValidationRule{
		Name: "numeric",
		Validator: func(value interface{}) error {
			switch v := value.(type) {
			case int, int8, int16, int32, int64:
				return nil
			case uint, uint8, uint16, uint32, uint64:
				return nil
			case float32, float64:
				return nil
			case string:
				if _, err := strconv.ParseFloat(v, 64); err != nil {
					return fmt.Errorf("value must be numeric")
				}
				return nil
			default:
				return fmt.Errorf("value must be numeric")
			}
		},
		Message: "%s must be numeric",
	}

	// Positive rule
	Positive = ValidationRule{
		Name: "positive",
		Validator: func(value interface{}) error {
			var num float64
			switch v := value.(type) {
			case int:
				num = float64(v)
			case int64:
				num = float64(v)
			case float64:
				num = v
			case string:
				parsed, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return fmt.Errorf("value must be numeric")
				}
				num = parsed
			default:
				return fmt.Errorf("value must be numeric")
			}

			if num <= 0 {
				return fmt.Errorf("value must be positive")
			}
			return nil
		},
		Message: "%s must be positive",
	}

	// Range rule
	Range = func(min, max float64) ValidationRule {
		return ValidationRule{
			Name: fmt.Sprintf("range_%.1f_%.1f", min, max),
			Validator: func(value interface{}) error {
				var num float64
				switch v := value.(type) {
				case int:
					num = float64(v)
				case int64:
					num = float64(v)
				case float64:
					num = v
				case string:
					parsed, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return fmt.Errorf("value must be numeric")
					}
					num = parsed
				default:
					return fmt.Errorf("value must be numeric")
				}

				if num < min || num > max {
					return fmt.Errorf("value must be between %.1f and %.1f", min, max)
				}
				return nil
			},
			Message: fmt.Sprintf("%%s must be between %.1f and %.1f", min, max),
		}
	}
)

// DefaultValidator returns validator with common rules
func DefaultValidator() *Validator {
	v := NewValidator()
	v.AddRule("required", Required.Validator, Required.Message)
	v.AddRule("email", Email.Validator, Email.Message)
	v.AddRule("numeric", Numeric.Validator, Numeric.Message)
	v.AddRule("positive", Positive.Validator, Positive.Message)
	return v
}

// ValidateStruct validates struct using default validator
func ValidateStruct(obj interface{}) ValidationErrors {
	v := DefaultValidator()
	return v.ValidateStruct(obj)
}

// CreateValidationError creates validation error
func CreateValidationError(field string, value interface{}, rule, message string) Error {
	metadata := Metadata{
		"field": field,
		"value": value,
		"rule":  rule,
		"validation_errors": map[string]interface{}{
			field: map[string]interface{}{
				"value":   value,
				"rule":    rule,
				"message": message,
			},
		},
	}

	return NewWithMetadata(ErrCodeValidationFailed,
		fmt.Sprintf("Validation failed for field %s", field), metadata)
}

// CreateValidationErrors creates error from multiple validation errors
func CreateValidationErrors(errors ValidationErrors) Error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		err := errors[0]
		return CreateValidationError(err.Field, err.Value, err.Rule, err.Message)
	}

	metadata := Metadata{
		"validation_errors": errors.ToMap(),
		"error_count":       len(errors),
	}

	return NewWithMetadata(ErrCodeValidationFailed,
		fmt.Sprintf("Validation failed with %d errors", len(errors)), metadata)
}

// FieldValidation provides field-level validation
type FieldValidation struct {
	Field string
	Value interface{}
	Rules []string
}

// ValidateFields validates multiple fields
func ValidateFields(validations []FieldValidation) ValidationErrors {
	v := DefaultValidator()
	var allErrors ValidationErrors

	for _, validation := range validations {
		errors := v.ValidateField(validation.Field, validation.Value, validation.Rules)
		allErrors = append(allErrors, errors...)
	}

	return allErrors
}

// ValidationBuilder provides fluent API for validation
type ValidationBuilder struct {
	validations []FieldValidation
}

// NewValidationBuilder creates new validation builder
func NewValidationBuilder() *ValidationBuilder {
	return &ValidationBuilder{
		validations: make([]FieldValidation, 0),
	}
}

// Field adds field validation
func (vb *ValidationBuilder) Field(field string, value interface{}) *FieldValidationBuilder {
	return &FieldValidationBuilder{
		builder: vb,
		field:   field,
		value:   value,
	}
}

// FieldValidationBuilder builds field validation
type FieldValidationBuilder struct {
	builder *ValidationBuilder
	field   string
	value   interface{}
}

// Rules adds validation rules
func (fvb *FieldValidationBuilder) Rules(rules ...string) *ValidationBuilder {
	fvb.builder.validations = append(fvb.builder.validations, FieldValidation{
		Field: fvb.field,
		Value: fvb.value,
		Rules: rules,
	})
	return fvb.builder
}

// Validate executes all validations
func (vb *ValidationBuilder) Validate() ValidationErrors {
	return ValidateFields(vb.validations)
}

// ValidateWithError executes validations and returns error
func (vb *ValidationBuilder) ValidateWithError() Error {
	errors := vb.Validate()
	if len(errors) > 0 {
		return CreateValidationErrors(errors)
	}
	return nil
}
