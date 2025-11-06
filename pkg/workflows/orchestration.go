package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/melslow/kitsune/pkg/activities/handlers"
	"github.com/melslow/kitsune/pkg/models"
)

// OrchestrationWorkflow coordinates execution across multiple servers
func OrchestrationWorkflow(ctx workflow.Context, req models.ExecutionRequest) (*models.OrchestrationResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting orchestration", "servers", len(req.Servers), "strategy", req.RolloutStrategy.Type)

	// Validate all steps before dispatching to workers
	validator := handlers.NewStepValidator()
	if err := validator.ValidateSteps(req.Steps); err != nil {
		logger.Error("Step validation failed", "error", err)
		return nil, fmt.Errorf("step validation failed: %w", err)
	}
	logger.Info("All steps validated successfully")

	result := &models.OrchestrationResult{
		Results: make([]models.ExecutionResult, 0),
	}

	var results []models.ExecutionResult
	var err error

	switch req.RolloutStrategy.Type {
	case "Parallel":
		results, err = parallelExecution(ctx, req)
	case "Rolling":
		results, err = rollingExecution(ctx, req)
	case "Sequential":
		results, err = sequentialExecution(ctx, req)
	default:
		results, err = parallelExecution(ctx, req)
	}

	if err != nil {
		return nil, err
	}

	// Count results
	for _, r := range results {
		result.Results = append(result.Results, r)
		if r.Success {
			result.ServersPatched++
		} else {
			result.ServersFailed++
		}
	}

	result.Success = result.ServersFailed == 0

	logger.Info("Orchestration complete", "success", result.Success, "patched", result.ServersPatched, "failed", result.ServersFailed)

	return result, nil
}

func parallelExecution(ctx workflow.Context, req models.ExecutionRequest) ([]models.ExecutionResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting parallel execution", "servers", len(req.Servers))

	var futures []workflow.Future

	for _, serverID := range req.Servers {
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("exec-%s", serverID),
			TaskQueue:  serverID,
		})

		input := models.WorkflowInput{
			ServerID: serverID,
			Steps:    req.Steps,
		}
		future := workflow.ExecuteChildWorkflow(childCtx, ServerExecutionWorkflow, input)
		futures = append(futures, future)
	}

	var results []models.ExecutionResult
	failures := 0
	
	for i, future := range futures {
		var result models.ExecutionResult
		if err := future.Get(ctx, &result); err != nil {
			result = models.ExecutionResult{
				ServerID: req.Servers[i],
				Success:  false,
				Error:    err.Error(),
			}
		}
		results = append(results, result)
		
		if !result.Success {
			failures++
		}
	}

	// Check if max failures exceeded and trigger rollback
	if req.RolloutStrategy.MaxFailures >= 0 && failures > req.RolloutStrategy.MaxFailures {
		logger.Error("Max failures exceeded, triggering rollback", "failures", failures, "maxFailures", req.RolloutStrategy.MaxFailures)
		
		// Trigger rollback on all servers that were processed
		for _, result := range results {
			if result.Success {
				logger.Info("Triggering rollback for server", "serverID", result.ServerID)
				if err := triggerServerRollback(ctx, result.ServerID, req.Steps, result); err != nil {
					logger.Error("Failed to trigger rollback", "serverID", result.ServerID, "error", err)
				}
			}
		}
		
		return results, fmt.Errorf("exceeded max failures: %d > %d", failures, req.RolloutStrategy.MaxFailures)
	}

	return results, nil
}

func sequentialExecution(ctx workflow.Context, req models.ExecutionRequest) ([]models.ExecutionResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting sequential execution", "servers", len(req.Servers))

	var results []models.ExecutionResult
	failures := 0

	for _, serverID := range req.Servers {
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("exec-%s", serverID),
			TaskQueue:  serverID,
		})

		input := models.WorkflowInput{
			ServerID: serverID,
			Steps:    req.Steps,
		}
		var result models.ExecutionResult
		err := workflow.ExecuteChildWorkflow(childCtx, ServerExecutionWorkflow, input).Get(ctx, &result)

		if err != nil {
			result = models.ExecutionResult{
				ServerID: serverID,
				Success:  false,
				Error:    err.Error(),
			}
		}

		results = append(results, result)

		if !result.Success {
			failures++
		}

		// Check if max failures exceeded
		if req.RolloutStrategy.MaxFailures >= 0 && failures > req.RolloutStrategy.MaxFailures {
			logger.Error("Max failures exceeded, triggering rollback", "failures", failures, "maxFailures", req.RolloutStrategy.MaxFailures)
			
			// Trigger rollback on all successfully executed servers
			for _, prevResult := range results {
				if prevResult.Success {
					logger.Info("Triggering rollback for server", "serverID", prevResult.ServerID)
					if err := triggerServerRollback(ctx, prevResult.ServerID, req.Steps, prevResult); err != nil {
						logger.Error("Failed to trigger rollback", "serverID", prevResult.ServerID, "error", err)
					}
				}
			}
			
			return results, fmt.Errorf("exceeded max failures: %d > %d", failures, req.RolloutStrategy.MaxFailures)
		}
	}

	return results, nil
}

func triggerServerRollback(ctx workflow.Context, serverID string, steps []models.StepDefinition, executionResult models.ExecutionResult) error {
	logger := workflow.GetLogger(ctx)
	
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("rollback-%s", serverID),
		TaskQueue:  serverID,
	})
	
	// Build executed steps info from the execution result
	var executedSteps []ExecutedStepInfo
	for i, stepResult := range executionResult.StepsExecuted {
		if stepResult.Success && i < len(steps) {
			executedSteps = append(executedSteps, ExecutedStepInfo{
				Step:     steps[i],
				Metadata: nil,
			})
		}
	}
	
	// Execute rollback as child workflow
	logger.Info("Starting rollback workflow", "serverID", serverID, "steps", len(executedSteps))
	
	input := RollbackWorkflowInput{
		ServerID:      serverID,
		ExecutedSteps: executedSteps,
	}
	
	err := workflow.ExecuteChildWorkflow(childCtx, ServerRollbackWorkflow, input).Get(ctx, nil)
	if err != nil {
		logger.Error("Rollback workflow failed", "serverID", serverID, "error", err)
		return err
	}
	
	logger.Info("Rollback workflow completed", "serverID", serverID)
	return nil
}

func rollingExecution(ctx workflow.Context, req models.ExecutionRequest) ([]models.ExecutionResult, error) {
	logger := workflow.GetLogger(ctx)

	batchSize := req.RolloutStrategy.BatchSize
	if batchSize == 0 {
		batchSize = 1
	}

	logger.Info("Starting rolling execution", "servers", len(req.Servers), "batchSize", batchSize)

	var allResults []models.ExecutionResult
	failures := 0

	for i := 0; i < len(req.Servers); i += batchSize {
		end := i + batchSize
		if end > len(req.Servers) {
			end = len(req.Servers)
		}
		batch := req.Servers[i:end]

		logger.Info("Processing batch", "batch", (i/batchSize)+1, "servers", batch)

		// Execute batch in parallel
		var futures []workflow.Future
		for _, serverID := range batch {
			childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
				WorkflowID: fmt.Sprintf("exec-%s", serverID),
				TaskQueue:  serverID,
			})

			input := models.WorkflowInput{
				ServerID: serverID,
				Steps:    req.Steps,
			}
			future := workflow.ExecuteChildWorkflow(childCtx, ServerExecutionWorkflow, input)
			futures = append(futures, future)
		}

		// Wait for batch
		for j, future := range futures {
			var result models.ExecutionResult
			if err := future.Get(ctx, &result); err != nil {
				result = models.ExecutionResult{
					ServerID: batch[j],
					Success:  false,
					Error:    err.Error(),
				}
			}

			allResults = append(allResults, result)

			if !result.Success {
				failures++
			}
		}

		// Check failure threshold
		if req.RolloutStrategy.MaxFailures >= 0 && failures > req.RolloutStrategy.MaxFailures {
			logger.Error("Max failures exceeded, triggering rollback", "failures", failures, "maxFailures", req.RolloutStrategy.MaxFailures)
			
			// Trigger rollback on all successfully executed servers
			for _, result := range allResults {
				if result.Success {
					logger.Info("Triggering rollback for server", "serverID", result.ServerID)
					if err := triggerServerRollback(ctx, result.ServerID, req.Steps, result); err != nil {
						logger.Error("Failed to trigger rollback", "serverID", result.ServerID, "error", err)
					}
				}
			}
			
			return allResults, fmt.Errorf("exceeded max failures: %d > %d", failures, req.RolloutStrategy.MaxFailures)
		}

		// Delay between batches
		if end < len(req.Servers) && req.RolloutStrategy.BatchDelaySeconds > 0 {
			workflow.Sleep(ctx, time.Duration(req.RolloutStrategy.BatchDelaySeconds)*time.Second)
		}
	}

	return allResults, nil
}
