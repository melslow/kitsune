package models

// StepDefinition represents a single step to execute
type StepDefinition struct {
	Name              string                 `json:"name"`
	Type              string                 `json:"type"`
	Params            map[string]interface{} `json:"params,omitempty"`
	Required          bool                   `json:"required"`
	ContinueOnFailure bool                   `json:"continueOnFailure"`
}

// ExecutionResult is the result of executing steps on one server
type ExecutionResult struct {
	ServerID      string       `json:"serverId"`
	Success       bool         `json:"success"`
	Error         string       `json:"error,omitempty"`
	StepsExecuted []StepResult `json:"stepsExecuted"`
}

// StepResult is the result of a single step
type StepResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}