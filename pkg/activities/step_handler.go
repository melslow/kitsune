package activities

import "context"

// ExecutionMetadata contains data captured during execution that may be needed for rollback
type ExecutionMetadata map[string]interface{}

// StepHandler defines the interface all step types must implement
type StepHandler interface {
	Execute(ctx context.Context, params map[string]interface{}) (ExecutionMetadata, error)
	Rollback(ctx context.Context, params map[string]interface{}, metadata ExecutionMetadata) error
}

// StepHandlerRegistry manages all registered step handlers
type StepHandlerRegistry struct {
	handlers map[string]StepHandler
}

func NewStepHandlerRegistry() *StepHandlerRegistry {
	return &StepHandlerRegistry{
		handlers: make(map[string]StepHandler),
	}
}

func (r *StepHandlerRegistry) Register(stepType string, handler StepHandler) {
	r.handlers[stepType] = handler
}

func (r *StepHandlerRegistry) Get(stepType string) (StepHandler, bool) {
	handler, ok := r.handlers[stepType]
	return handler, ok
}