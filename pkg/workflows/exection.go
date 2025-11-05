package workflows

import (
	"fmt"
	"time"
	
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	
	"github.com/melslow/kitsune/pkg/models"
)

// ServerExecutionWorkflow executes a list of steps on a single server
func ServerExecutionWorkflow(ctx workflow.Context, serverID string, steps []models.StepDefinition) (models.ExecutionResult, error) {
	logger := workflow.GetLogger(ctx)
	result := models.ExecutionResult{
		ServerID:      serverID,
		StepsExecuted: []models.StepResult{},
	}
	
	logger.Info("Starting execution workflow", "serverID", serverID, "steps", len(steps))
	
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
	
	var executedSteps []models.StepDefinition
	
	// Execute each step
	for i, step := range steps {
		logger.Info("Executing step", "number", i+1, "name", step.Name, "type", step.Type)
		
		err := workflow.ExecuteActivity(ctx, "ExecuteStep", serverID, step).Get(ctx, nil)
		
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
				
				// Rollback
				rollbackSteps(ctx, serverID, executedSteps)
				return result, err
			}
			
			logger.Warn("Step failed but continuing", "step", step.Name)
		} else {
			stepResult.Success = true
			executedSteps = append(executedSteps, step)
		}
		
		result.StepsExecuted = append(result.StepsExecuted, stepResult)
	}
	
	result.Success = true
	logger.Info("Execution workflow completed", "serverID", serverID)
	
	return result, nil
}

func rollbackSteps(ctx workflow.Context, serverID string, steps []models.StepDefinition) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Rolling back steps", "count", len(steps))
	
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		logger.Info("Rolling back step", "step", step.Name)
		workflow.ExecuteActivity(ctx, "RollbackStep", serverID, step).Get(ctx, nil)
	}
}