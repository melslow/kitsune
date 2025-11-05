package workflows

import (
	"fmt"
	"time"
	
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	
	"github.com/melslow/kitsune/pkg/models"
)

type ExecutedStepInfo struct {
	Step     models.StepDefinition
	Metadata map[string]interface{}
}

type RollbackWorkflowInput struct {
	ServerID      string
	ExecutedSteps []ExecutedStepInfo
}

// ServerExecutionWorkflow executes a list of steps on a single server
func ServerExecutionWorkflow(ctx workflow.Context, input models.WorkflowInput) (models.ExecutionResult, error) {
	logger := workflow.GetLogger(ctx)
	result := models.ExecutionResult{
		ServerID:      input.ServerID,
		StepsExecuted: []models.StepResult{},
	}
	
	logger.Info("Starting execution workflow", "serverID", input.ServerID, "steps", len(input.Steps))
	
	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	// Execute each step
	for i, step := range input.Steps {
		logger.Info("Executing step", "number", i+1, "name", step.Name, "type", step.Type)
		
		var metadata map[string]interface{}
		err := workflow.ExecuteActivity(ctx, "ExecuteStep", input.ServerID, step).Get(ctx, &metadata)
		
		stepResult := models.StepResult{
			Name: step.Name,
		}
		
		if err != nil {
			stepResult.Success = false
			stepResult.Error = err.Error()
			
			if step.Required && !step.ContinueOnFailure {
				logger.Error("Required step failed", "step", step.Name, "error", err)
				result.Error = fmt.Sprintf("Required step '%s' failed: %v", step.Name, err)
				result.StepsExecuted = append(result.StepsExecuted, stepResult)
				return result, err
			}
			
			logger.Warn("Step failed but continuing", "step", step.Name)
		} else {
			stepResult.Success = true
		}
		
		result.StepsExecuted = append(result.StepsExecuted, stepResult)
	}
	
	result.Success = true
	logger.Info("Execution workflow completed", "serverID", input.ServerID)
	
	return result, nil
}

// ServerRollbackWorkflow executes rollback steps for a server
func ServerRollbackWorkflow(ctx workflow.Context, input RollbackWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting rollback workflow", "serverID", input.ServerID, "steps", len(input.ExecutedSteps))
	
	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	rollbackSteps(ctx, input.ServerID, input.ExecutedSteps)
	
	logger.Info("Rollback workflow completed", "serverID", input.ServerID)
	return nil
}

func rollbackSteps(ctx workflow.Context, serverID string, steps []ExecutedStepInfo) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Rolling back steps", "count", len(steps))
	
	for i := len(steps) - 1; i >= 0; i-- {
		stepInfo := steps[i]
		logger.Info("Rolling back step", "step", stepInfo.Step.Name)
		workflow.ExecuteActivity(ctx, "RollbackStep", serverID, stepInfo.Step, stepInfo.Metadata).Get(ctx, nil)
	}
}