package params

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ParseAndValidate parses raw parameters into a typed struct and validates that no unsupported parameters are present
func ParseAndValidate(raw map[string]interface{}, target interface{}) error {
	if raw == nil {
		return fmt.Errorf("parameters cannot be nil")
	}

	// Get supported parameter names from struct tags
	supportedParams := getSupportedParams(target)

	// Check for unsupported parameters
	var unsupportedParams []string
	for key := range raw {
		if !supportedParams[key] {
			unsupportedParams = append(unsupportedParams, key)
		}
	}

	if len(unsupportedParams) > 0 {
		return fmt.Errorf("unsupported parameters: %s", strings.Join(unsupportedParams, ", "))
	}

	// Parse parameters into target struct
	jsonData, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate required fields
	if err := validateRequired(target); err != nil {
		return err
	}

	return nil
}

// getSupportedParams extracts parameter names from struct json tags
func getSupportedParams(target interface{}) map[string]bool {
	supported := make(map[string]bool)
	
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Extract just the name part (before comma)
			name := strings.Split(jsonTag, ",")[0]
			supported[name] = true
		}
	}
	
	return supported
}

// validateRequired checks that all required fields are set
func validateRequired(target interface{}) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		validateTag := field.Tag.Get("validate")
		
		if strings.Contains(validateTag, "required") {
			fieldValue := v.Field(i)
			
			// Check if field is zero value
			if isZeroValue(fieldValue) {
				jsonTag := field.Tag.Get("json")
				name := strings.Split(jsonTag, ",")[0]
				if name == "" {
					name = field.Name
				}
				return fmt.Errorf("missing required parameter: %s", name)
			}
		}
	}
	
	return nil
}

// isZeroValue checks if a reflect.Value is the zero value for its type
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
