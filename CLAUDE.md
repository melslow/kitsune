# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Kitsune is a distributed orchestration system for executing workflows across multiple servers using Temporal. It enables coordinated deployments, patches, and operations across server fleets with flexible rollout strategies (Parallel, Sequential, Rolling).

## Common Commands

### Build
```bash
go build ./cmd/local-worker
go build ./cmd/orchestration-worker
```

### Run Tests
```bash
go test ./...
```

### Run Specific Test
```bash
go test ./pkg/activities/handlers -v -run TestName
```

### Local Development
```bash
cd dev
docker-compose up -d     # Start Temporal + workers
docker-compose logs -f   # View logs
docker-compose down      # Stop all services
```

### Trigger Test Workflow
```bash
temporal workflow start \
  --task-queue execution-orchestrator \
  --type OrchestrationWorkflow \
  --input @dev/test-plan-orchestrator.json
```

### View Workflow Status
```bash
temporal workflow describe --workflow-id <workflow-id>
```

## Architecture

### Hybrid Worker Model
Kitsune uses two types of Temporal workers:

1. **Central Orchestrator Worker** (`cmd/orchestration-worker/`)
   - Runs `OrchestrationWorkflow` on `execution-orchestrator` task queue
   - Coordinates execution across servers
   - Handles rollout strategies and failure thresholds
   - Triggers rollbacks when maxFailures exceeded

2. **Local Workers** (`cmd/local-worker/`)
   - One worker per target server
   - Listens on server-specific task queue (e.g., `server-1`, `server-2`)
   - Runs `ServerExecutionWorkflow` and `ServerRollbackWorkflow`
   - Executes actual step handlers

### Workflow Flow
1. OrchestrationWorkflow receives ExecutionRequest (servers, steps, rollout strategy)
2. Validates all steps upfront using StepValidator
3. Spawns ServerExecutionWorkflow as child workflow for each server (routed via task queue)
4. Each ServerExecutionWorkflow validates and executes steps sequentially
5. If maxFailures exceeded, OrchestrationWorkflow triggers ServerRollbackWorkflow for successful servers

### Key Design Patterns

**Task Queue Routing**: Child workflows are routed to specific workers via task queue names matching server IDs:
```go
childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
    WorkflowID: fmt.Sprintf("exec-%s", serverID),
    TaskQueue:  serverID,  // Routes to worker listening on this queue
})
```

**Two-Level Validation**: Steps are validated at both workflow level (early detection) and handler level (runtime safety) using `params.ParseAndValidate()`.

**Rollback Mechanism**: When a required step fails or maxFailures is exceeded, Kitsune:
1. Stops new server executions
2. Spawns ServerRollbackWorkflow for each successfully executed server
3. Calls handler.Rollback() for each executed step in reverse order

## Core Components

### Models (`pkg/models/types.go`)
- `ExecutionRequest`: Input for OrchestrationWorkflow
- `WorkflowInput`: Input for ServerExecutionWorkflow
- `StepDefinition`: Defines a single step (name, type, params, required, continueOnFailure)
- `RolloutStrategy`: Controls execution flow (type, batchSize, batchDelaySeconds, maxFailures)
- `ExecutionResult`/`StepResult`: Capture execution outcomes

### Workflows (`pkg/workflows/`)
- `OrchestrationWorkflow`: Coordinates across servers with rollout strategies
- `ServerExecutionWorkflow`: Executes steps on single server
- `ServerRollbackWorkflow`: Rolls back executed steps on single server

### Activities (`pkg/activities/`)
- `StepActivities`: Wraps handler execution
  - `ExecuteStep`: Dispatches to handler based on step.Type
  - `RollbackStep`: Dispatches rollback to handler
- `StepHandlerRegistry`: Maps step types to handlers

### Handlers (`pkg/activities/handlers/`)
All handlers implement the `StepHandler` interface:
```go
type StepHandler interface {
    Execute(ctx context.Context, params map[string]interface{}) (ExecutionMetadata, error)
    Rollback(ctx context.Context, params map[string]interface{}, metadata ExecutionMetadata) error
}
```

Built-in handlers:
- `EchoHandler`: Logs messages
- `ScriptHandler`: Executes shell scripts with optional rollback script
- `SleepHandler`: Adds delays
- `FileWriteHandler`: Writes files to disk
- `YumUpgradeHandler`: Upgrades yum packages with automatic rollback

### Validation (`pkg/activities/handlers/validator.go` and `pkg/activities/params/`)
- Each handler defines a typed params struct with `json` and `validate:"required"` tags
- `params.ParseAndValidate()` rejects unsupported parameters and validates required fields
- `StepValidator.ValidateSteps()` validates all steps before execution starts
- See `docs/handler_params_validation.md` for detailed validation documentation

## Adding a New Handler

1. Create handler in `pkg/activities/handlers/`:
```go
type CustomParams struct {
    Field string `json:"field" validate:"required"`
}

type CustomHandler struct{}

func (h *CustomHandler) Execute(ctx context.Context, params map[string]interface{}) (activities.ExecutionMetadata, error) {
    var p CustomParams
    if err := params.ParseAndValidate(params, &p); err != nil {
        return nil, err
    }
    // Execute logic
    return activities.ExecutionMetadata{"key": "value"}, nil
}

func (h *CustomHandler) Rollback(ctx context.Context, params map[string]interface{}, metadata activities.ExecutionMetadata) error {
    // Rollback logic
    return nil
}
```

2. Register in `cmd/local-worker/main.go`:
```go
registry.Register("custom", &handlers.CustomHandler{})
```

3. Add to validator in `pkg/activities/handlers/validator.go`:
```go
case "custom":
    return &CustomParams{}
```

## Configuration

### Environment Variables
- `SERVER_ID`: Server identifier for local worker (task queue name)
- `TEMPORAL_ADDRESS`: Temporal server address (default: `localhost:7233`)

### Rollout Strategies
- **Parallel**: Execute on all servers simultaneously
- **Sequential**: One server at a time, stop on first failure if maxFailures=0
- **Rolling**: Batches with configurable size and delays

### Error Handling
- `required: true`: Step failure triggers workflow failure and rollback
- `continueOnFailure: true`: Step failure is logged but execution continues
- `maxFailures`: Number of server failures allowed before stopping rollout and triggering rollback (0 = stop on first failure)

## Development Notes

- Temporal UI runs on http://localhost:8080 in dev environment
- Use `temporal` CLI to start workflows and inspect executions
- Handler params are validated at workflow level before execution and at handler level during execution
- Rollback is automatic when maxFailures threshold is exceeded
- Each handler should use `params.ParseAndValidate()` to enforce parameter schemas
- Server-specific routing is handled by Temporal task queues matching SERVER_ID
