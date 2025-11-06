package handlers

import (
	"fmt"
	
	"github.com/melslow/kitsune/pkg/activities/params"
	"github.com/melslow/kitsune/pkg/models"
)

// StepValidator validates step parameters before execution
type StepValidator struct{}

func NewStepValidator() *StepValidator {
	return &StepValidator{}
}

// ValidateStep validates parameters for a single step
func (v *StepValidator) ValidateStep(step models.StepDefinition) error {
	// Get the appropriate params struct for the step type
	paramsStruct := v.getParamsStructForType(step.Type)
	if paramsStruct == nil {
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
	
	// Validate the parameters
	if err := params.ParseAndValidate(step.Params, paramsStruct); err != nil {
		return fmt.Errorf("validation failed for step '%s' (type: %s): %w", step.Name, step.Type, err)
	}
	
	return nil
}

// ValidateSteps validates all steps in a list
func (v *StepValidator) ValidateSteps(steps []models.StepDefinition) error {
	for i, step := range steps {
		if err := v.ValidateStep(step); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return nil
}

// getParamsStructForType returns an empty params struct for the given step type
func (v *StepValidator) getParamsStructForType(stepType string) interface{} {
	switch stepType {
	case "echo":
		return &EchoParams{}
	case "sleep":
		return &SleepParams{}
	case "file_write":
		return &FileWriteParams{}
	case "script":
		return &ScriptParams{}
	case "yum_upgrade":
		return &YumUpgradeParams{}
	default:
		return nil
	}
}
