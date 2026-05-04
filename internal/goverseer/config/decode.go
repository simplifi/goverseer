package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = newValidator()

// Decode unmarshals a dynamic config map into a typed config struct and then
// validates the struct using validator tags.
func Decode(input interface{}, output interface{}) error {
	cfgMap, err := configMap(input)
	if err != nil {
		return err
	}

	v := viper.New()
	if err := v.MergeConfigMap(cfgMap); err != nil {
		return err
	}
	if err := checkConfigFields(cfgMap, output); err != nil {
		return err
	}
	if err := v.UnmarshalExact(output); err != nil {
		return err
	}

	return Validate(output)
}

// Validate runs configured validation rules against a typed config struct.
func Validate(output interface{}) error {
	if err := validate.Struct(output); err != nil {
		return formatValidationError(err)
	}

	return nil
}

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := mapstructureName(field)
		if name == "-" {
			return ""
		}
		if name == "" {
			return field.Name
		}
		return name
	})
	return v
}

func mapstructureName(field reflect.StructField) string {
	return strings.SplitN(field.Tag.Get("mapstructure"), ",", 2)[0]
}

func configMap(input interface{}) (map[string]interface{}, error) {
	if input == nil {
		return map[string]interface{}{}, nil
	}

	normalized, err := normalizeValue(input)
	if err != nil {
		return nil, err
	}

	cfgMap, ok := normalized.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	return cfgMap, nil
}

func normalizeValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch typed := value.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(typed))
		for key, nestedValue := range typed {
			value, err := normalizeValue(nestedValue)
			if err != nil {
				return nil, err
			}
			if value != nil {
				normalized[key] = value
			}
		}
		return normalized, nil
	case map[interface{}]interface{}:
		normalized := make(map[string]interface{}, len(typed))
		for key, nestedValue := range typed {
			keyString, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("config keys must be strings")
			}

			value, err := normalizeValue(nestedValue)
			if err != nil {
				return nil, err
			}
			if value != nil {
				normalized[keyString] = value
			}
		}
		return normalized, nil
	case []interface{}:
		normalized := make([]interface{}, 0, len(typed))
		for _, nestedValue := range typed {
			value, err := normalizeValue(nestedValue)
			if err != nil {
				return nil, err
			}
			normalized = append(normalized, value)
		}
		return normalized, nil
	default:
		return value, nil
	}
}

func checkConfigFields(cfgMap map[string]interface{}, output interface{}) error {
	return checkStructConfigFields(cfgMap, output, "")
}

func checkStructConfigFields(cfgMap map[string]interface{}, output interface{}, prefix string) error {
	value := reflect.ValueOf(output)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return nil
	}

	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldName := mapstructureName(field)
		if fieldName == "" || fieldName == "-" {
			fieldName = field.Name
		}
		displayName := prefixedName(prefix, fieldName)

		rawValue, ok := cfgMap[fieldName]
		if !ok {
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			nestedMap, ok := rawValue.(map[string]interface{})
			if !ok {
				continue
			}
			if err := checkStructConfigFields(nestedMap, value.Field(i).Addr().Interface(), displayName); err != nil {
				return err
			}
			continue
		}

		if err := checkFieldTypeAndValue(field, rawValue, displayName); err != nil {
			return err
		}
	}

	return nil
}

func prefixedName(prefix string, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func checkFieldTypeAndValue(field reflect.StructField, rawValue interface{}, fieldName string) error {
	switch field.Type.Kind() {
	case reflect.String:
		stringValue, ok := rawValue.(string)
		if !ok {
			return fmt.Errorf("%s must be a string", fieldName)
		}
		if stringValue == "" && hasValidationTag(field.Tag.Get("validate"), "required") {
			return fmt.Errorf("%s must not be empty", fieldName)
		}
	case reflect.Int:
		if _, ok := rawValue.(int); !ok {
			return fmt.Errorf("%s must be an integer", fieldName)
		}
	case reflect.Bool:
		if _, ok := rawValue.(bool); !ok {
			return fmt.Errorf("%s must be a boolean", fieldName)
		}
	}

	return nil
}

func hasValidationTag(tag string, name string) bool {
	for _, rule := range strings.Split(tag, ",") {
		ruleName := strings.SplitN(rule, "=", 2)[0]
		if ruleName == name {
			return true
		}
	}
	return false
}

func formatValidationError(err error) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return err
	}

	fieldError := validationErrors[0]
	fieldName := validationFieldName(fieldError)

	switch fieldError.Tag() {
	case "required":
		return fmt.Errorf("%s is required", fieldName)
	case "gte":
		if fieldError.Param() == "1" {
			return fmt.Errorf("%s must be a positive integer", fieldName)
		}
		return fmt.Errorf("%s must be greater than or equal to %s", fieldName, fieldError.Param())
	case "gt":
		return fmt.Errorf("%s must be greater than %s", fieldName, fieldError.Param())
	case "oneof":
		return fmt.Errorf("%s must be one of %s", fieldName, fieldError.Param())
	default:
		return fmt.Errorf("%s failed validation for %s", fieldName, fieldError.Tag())
	}
}

func validationFieldName(fieldError validator.FieldError) string {
	namespace := fieldError.Namespace()
	if namespace == "" {
		return fieldError.Field()
	}

	parts := strings.Split(namespace, ".")
	if len(parts) <= 1 {
		return fieldError.Field()
	}

	return strings.Join(parts[1:], ".")
}
