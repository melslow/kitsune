package activities

import "context"

// StepHandler defines the interface all step types must implement
type StepHandler interface {
	Execute(ctx context.Context, params map[string]interface{}) error
	Rollback(ctx context.Context, params map[string]interface{}) error
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