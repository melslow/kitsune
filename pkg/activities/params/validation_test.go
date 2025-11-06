package params

import (
	"testing"
)

type TestParams struct {
	Required string `json:"required" validate:"required"`
	Optional string `json:"optional,omitempty"`
}

func TestParseAndValidate_Success(t *testing.T) {
	raw := map[string]interface{}{
		"required": "value",
		"optional": "optional_value",
	}

	var p TestParams
	err := ParseAndValidate(raw, &p)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if p.Required != "value" {
		t.Errorf("Expected required='value', got '%s'", p.Required)
	}

	if p.Optional != "optional_value" {
		t.Errorf("Expected optional='optional_value', got '%s'", p.Optional)
	}
}

func TestParseAndValidate_MissingRequired(t *testing.T) {
	raw := map[string]interface{}{
		"optional": "value",
	}

	var p TestParams
	err := ParseAndValidate(raw, &p)
	if err == nil {
		t.Error("Expected error for missing required parameter")
	}

	if err.Error() != "missing required parameter: required" {
		t.Errorf("Expected 'missing required parameter: required', got: %v", err)
	}
}

func TestParseAndValidate_UnsupportedParameter(t *testing.T) {
	raw := map[string]interface{}{
		"required":    "value",
		"unsupported": "bad",
	}

	var p TestParams
	err := ParseAndValidate(raw, &p)
	if err == nil {
		t.Error("Expected error for unsupported parameter")
	}

	if err.Error() != "unsupported parameters: unsupported" {
		t.Errorf("Expected 'unsupported parameters: unsupported', got: %v", err)
	}
}

func TestParseAndValidate_MultipleUnsupportedParameters(t *testing.T) {
	raw := map[string]interface{}{
		"required": "value",
		"bad1":     "value1",
		"bad2":     "value2",
	}

	var p TestParams
	err := ParseAndValidate(raw, &p)
	if err == nil {
		t.Error("Expected error for unsupported parameters")
	}

	// Should contain both unsupported params
	errMsg := err.Error()
	if errMsg != "unsupported parameters: bad1, bad2" && errMsg != "unsupported parameters: bad2, bad1" {
		t.Errorf("Expected unsupported parameters error with both bad1 and bad2, got: %v", err)
	}
}

func TestParseAndValidate_NilParams(t *testing.T) {
	var p TestParams
	err := ParseAndValidate(nil, &p)
	if err == nil {
		t.Error("Expected error for nil parameters")
	}

	if err.Error() != "parameters cannot be nil" {
		t.Errorf("Expected 'parameters cannot be nil', got: %v", err)
	}
}
