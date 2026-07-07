package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
)

var validate = newValidator()

// Decode unmarshals a dynamic config map into a typed config struct and then
// validates the struct using validator tags.
func Decode(input interface{}, output interface{}) error {
	cfgMap, err := configMap(input)
	if err != nil {
		return err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:     "mapstructure",
		Result:      output,
		ErrorUnused: true,
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(cfgMap); err != nil {
		return formatDecodeError(err)
	}

	return formatValidationErrorForInput(output, cfgMap)
}

// Validate runs configured validation rules against a typed config struct.
func Validate(output interface{}) error {
	if err := validate.Struct(output); err != nil {
		return formatValidationError(err, nil)
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

func formatValidationErrorForInput(output interface{}, cfgMap map[string]interface{}) error {
	if err := validate.Struct(output); err != nil {
		return formatValidationError(err, cfgMap)
	}

	return nil
}

func formatValidationError(err error, cfgMap map[string]interface{}) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return err
	}

	fieldError := validationErrors[0]
	fieldName := validationFieldName(fieldError)

	switch fieldError.Tag() {
	case "required":
		if fieldError.Kind() == reflect.String {
			if cfgMap != nil && !configPathExists(cfgMap, fieldName) {
				return fmt.Errorf("%s is required", fieldName)
			}
			return fmt.Errorf("%s must not be empty", fieldName)
		}
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

func configPathExists(cfgMap map[string]interface{}, path string) bool {
	if path == "" {
		return false
	}

	current := interface{}(cfgMap)
	for _, part := range strings.Split(path, ".") {
		nextMap, ok := current.(map[string]interface{})
		if !ok {
			return false
		}

		value, ok := nextMap[part]
		if !ok {
			return false
		}

		current = value
	}

	return true
}

func formatDecodeError(err error) error {
	message := err.Error()
	const prefix = "decoding failed due to the following error(s):\n\n"
	if strings.HasPrefix(message, prefix) {
		message = strings.TrimPrefix(message, prefix)
	}

	lines := strings.Split(message, "\n")
	formatted := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if message, ok := formatDecodeLine(line); ok {
			formatted = append(formatted, message)
			continue
		}

		return err
	}

	if len(formatted) == 0 {
		return err
	}

	if len(formatted) == 1 {
		return errors.New(formatted[0])
	}

	return errors.New(strings.Join(formatted, "; "))
}

func formatDecodeLine(line string) (string, bool) {
	if !strings.HasPrefix(line, "'") {
		return "", false
	}

	fieldEnd := strings.Index(line[1:], "'")
	if fieldEnd < 0 {
		return "", false
	}

	fieldName := line[1 : fieldEnd+1]
	rest := strings.TrimSpace(line[fieldEnd+2:])
	const expectedPrefix = "expected type '"
	if !strings.HasPrefix(rest, expectedPrefix) {
		return "", false
	}

	rest = strings.TrimPrefix(rest, expectedPrefix)
	typeEnd := strings.Index(rest, "'")
	if typeEnd < 0 {
		return "", false
	}

	expectedType := rest[:typeEnd]
	switch expectedType {
	case "string":
		return fmt.Sprintf("%s must be a string", fieldName), true
	case "bool":
		return fmt.Sprintf("%s must be a boolean", fieldName), true
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return fmt.Sprintf("%s must be an integer", fieldName), true
	default:
		return fmt.Sprintf("%s must be a %s", fieldName, expectedType), true
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
