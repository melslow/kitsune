package models

// WorkflowInput is the input for ServerExecutionWorkflow
type WorkflowInput struct {
	ServerID string           `json:"serverID"`
	Steps    []StepDefinition `json:"steps"`
}

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

// RolloutStrategy defines how to execute across servers
type RolloutStrategy struct {
	Type              string `json:"type"` // Rolling, Parallel, Sequential, Canary
	BatchSize         int    `json:"batchSize,omitempty"`
	BatchDelaySeconds int    `json:"batchDelaySeconds,omitempty"`
	MaxFailures       int    `json:"maxFailures,omitempty"`
	CanaryPercentage  int    `json:"canaryPercentage,omitempty"`
}

// ExecutionRequest is input for orchestration workflow
type ExecutionRequest struct {
	Servers         []string         `json:"servers"`
	Steps           []StepDefinition `json:"steps"`
	RolloutStrategy RolloutStrategy  `json:"rolloutStrategy"`
}

// OrchestrationResult is the output for orchestration workflow
type OrchestrationResult struct {
	Success        bool              `json:"success"`
	ServersPatched int               `json:"serversPatched"`
	ServersFailed  int               `json:"serversFailed"`
	Results        []ExecutionResult `json:"results"`
}
