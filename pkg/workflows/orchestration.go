package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/melslow/kitsune/pkg/models"
)

// OrchestrationWorkflow coordinates execution across multiple servers
func OrchestrationWorkflow(ctx workflow.Context, req models.ExecutionRequest) (*models.OrchestrationResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting orchestration", "servers", len(req.Servers), "strategy", req.RolloutStrategy.Type)

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
	}

	return results, nil
}

func sequentialExecution(ctx workflow.Context, req models.ExecutionRequest) ([]models.ExecutionResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting sequential execution", "servers", len(req.Servers))

	var results []models.ExecutionResult

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

		// Stop on failure if maxFailures is 0
		if !result.Success && req.RolloutStrategy.MaxFailures == 0 {
			logger.Error("Sequential execution stopped due to failure", "server", serverID)
			break
		}
	}

	return results, nil
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
		if req.RolloutStrategy.MaxFailures > 0 && failures > req.RolloutStrategy.MaxFailures {
			return allResults, fmt.Errorf("exceeded max failures: %d", failures)
		}

		// Delay between batches
		if end < len(req.Servers) && req.RolloutStrategy.BatchDelaySeconds > 0 {
			workflow.Sleep(ctx, time.Duration(req.RolloutStrategy.BatchDelaySeconds)*time.Second)
		}
	}

	return allResults, nil
}
