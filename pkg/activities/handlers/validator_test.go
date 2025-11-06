package handlers

import (
	"strings"
	"testing"
	
	"github.com/melslow/kitsune/pkg/models"
)

func TestStepValidator_ValidateStep_Success(t *testing.T) {
	validator := NewStepValidator()
	
	tests := []struct {
		name string
		step models.StepDefinition
	}{
		{
			name: "echo with valid params",
			step: models.StepDefinition{
				Name: "test echo",
				Type: "echo",
				Params: map[string]interface{}{
					"message": "hello",
				},
			},
		},
		{
			name: "sleep with valid params",
			step: models.StepDefinition{
				Name: "test sleep",
				Type: "sleep",
				Params: map[string]interface{}{
					"duration": 5.0,
				},
			},
		},
		{
			name: "file_write with valid params",
			step: models.StepDefinition{
				Name: "test file",
				Type: "file_write",
				Params: map[string]interface{}{
					"path":    "/tmp/test",
					"content": "data",
				},
			},
		},
		{
			name: "script with required params",
			step: models.StepDefinition{
				Name: "test script",
				Type: "script",
				Params: map[string]interface{}{
					"script": "/bin/echo",
				},
			},
		},
		{
			name: "script with all params",
			step: models.StepDefinition{
				Name: "test script full",
				Type: "script",
				Params: map[string]interface{}{
					"script":          "/bin/echo",
					"args":            []string{"hello"},
					"rollback_script": "/bin/true",
				},
			},
		},
		{
			name: "yum_upgrade with valid params",
			step: models.StepDefinition{
				Name: "test yum",
				Type: "yum_upgrade",
				Params: map[string]interface{}{
					"package": "nginx",
					"version": "1.20.0",
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStep(tt.step)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestStepValidator_ValidateStep_MissingRequired(t *testing.T) {
	validator := NewStepValidator()
	
	tests := []struct {
		name          string
		step          models.StepDefinition
		expectedError string
	}{
		{
			name: "echo missing message",
			step: models.StepDefinition{
				Name:   "test",
				Type:   "echo",
				Params: map[string]interface{}{},
			},
			expectedError: "missing required parameter: message",
		},
		{
			name: "sleep missing duration",
			step: models.StepDefinition{
				Name:   "test",
				Type:   "sleep",
				Params: map[string]interface{}{},
			},
			expectedError: "missing required parameter: duration",
		},
		{
			name: "file_write missing path",
			step: models.StepDefinition{
				Name: "test",
				Type: "file_write",
				Params: map[string]interface{}{
					"content": "data",
				},
			},
			expectedError: "missing required parameter: path",
		},
		{
			name: "yum_upgrade missing version",
			step: models.StepDefinition{
				Name: "test",
				Type: "yum_upgrade",
				Params: map[string]interface{}{
					"package": "nginx",
				},
			},
			expectedError: "missing required parameter: version",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStep(tt.step)
			if err == nil {
				t.Error("Expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestStepValidator_ValidateStep_UnsupportedParams(t *testing.T) {
	validator := NewStepValidator()
	
	tests := []struct {
		name          string
		step          models.StepDefinition
		expectedError string
	}{
		{
			name: "echo with unsupported param",
			step: models.StepDefinition{
				Name: "test",
				Type: "echo",
				Params: map[string]interface{}{
					"message":     "hello",
					"unsupported": "value",
				},
			},
			expectedError: "unsupported parameters: unsupported",
		},
		{
			name: "yum_upgrade with typo",
			step: models.StepDefinition{
				Name: "test",
				Type: "yum_upgrade",
				Params: map[string]interface{}{
					"package": "nginx",
					"version": "1.20.0",
					"verison": "typo",
				},
			},
			expectedError: "unsupported parameters: verison",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStep(tt.step)
			if err == nil {
				t.Error("Expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestStepValidator_ValidateStep_UnknownType(t *testing.T) {
	validator := NewStepValidator()
	
	step := models.StepDefinition{
		Name:   "test",
		Type:   "unknown_handler",
		Params: map[string]interface{}{},
	}
	
	err := validator.ValidateStep(step)
	if err == nil {
		t.Error("Expected error for unknown step type")
	}
	
	if !strings.Contains(err.Error(), "unknown step type") {
		t.Errorf("Expected 'unknown step type' error, got: %v", err)
	}
}

func TestStepValidator_ValidateSteps_Multiple(t *testing.T) {
	validator := NewStepValidator()
	
	steps := []models.StepDefinition{
		{
			Name: "step 1",
			Type: "echo",
			Params: map[string]interface{}{
				"message": "hello",
			},
		},
		{
			Name: "step 2",
			Type: "sleep",
			Params: map[string]interface{}{
				"duration": 1.0,
			},
		},
		{
			Name: "step 3",
			Type: "yum_upgrade",
			Params: map[string]interface{}{
				"package": "nginx",
				"version": "1.20.0",
			},
		},
	}
	
	err := validator.ValidateSteps(steps)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestStepValidator_ValidateSteps_ErrorInSecondStep(t *testing.T) {
	validator := NewStepValidator()
	
	steps := []models.StepDefinition{
		{
			Name: "step 1",
			Type: "echo",
			Params: map[string]interface{}{
				"message": "hello",
			},
		},
		{
			Name: "step 2 - invalid",
			Type: "sleep",
			Params: map[string]interface{}{
				"duration":    1.0,
				"unsupported": "param",
			},
		},
		{
			Name: "step 3",
			Type: "yum_upgrade",
			Params: map[string]interface{}{
				"package": "nginx",
				"version": "1.20.0",
			},
		},
	}
	
	err := validator.ValidateSteps(steps)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	if !strings.Contains(err.Error(), "step 2") {
		t.Errorf("Expected error to indicate step 2, got: %v", err)
	}
	
	if !strings.Contains(err.Error(), "unsupported parameters") {
		t.Errorf("Expected unsupported parameters error, got: %v", err)
	}
}
