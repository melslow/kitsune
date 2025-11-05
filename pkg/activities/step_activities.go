package activities

import (
	"context"
	"fmt"
	
	"go.temporal.io/sdk/activity"
	
	"github.com/melslow/kitsune/pkg/models"
)

type StepActivities struct {
	serverID string
	registry *StepHandlerRegistry
}

func NewStepActivities(serverID string, registry *StepHandlerRegistry) *StepActivities {
	return &StepActivities{
		serverID: serverID,
		registry: registry,
	}
}

// ExecuteStep executes a single step using the handler registry
func (a *StepActivities) ExecuteStep(ctx context.Context, serverID string, step models.StepDefinition) (ExecutionMetadata, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing step", "name", step.Name, "type", step.Type)
	
	handler, ok := a.registry.Get(step.Type)
	if !ok {
		return nil, fmt.Errorf("no handler registered for step type: %s", step.Type)
	}
	
	// Add serverID to params
	if step.Params == nil {
		step.Params = make(map[string]interface{})
	}
	step.Params["server_id"] = serverID
	
	return handler.Execute(ctx, step.Params)
}

// RollbackStep rolls back a step
func (a *StepActivities) RollbackStep(ctx context.Context, serverID string, step models.StepDefinition, metadata ExecutionMetadata) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Rolling back step", "name", step.Name, "type", step.Type)
	
	handler, ok := a.registry.Get(step.Type)
	if !ok {
		logger.Warn("No handler for rollback", "type", step.Type)
		return nil
	}
	
	if step.Params == nil {
		step.Params = make(map[string]interface{})
	}
	step.Params["server_id"] = serverID
	
	return handler.Rollback(ctx, step.Params, metadata)
}