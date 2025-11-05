<p align="center">
  <img src="docs/img/logo.png" alt="Kitsune Logo" width="400" style="border-radius: 50%;"/>
</p>

# Kitsune

A distributed orchestration system for executing workflows across multiple servers using [Temporal](https://temporal.io/). Kitsune enables coordinated deployments, patches, and operations across server fleets with flexible rollout strategies.

## Architecture

Kitsune uses a hybrid architecture with two types of workers:

### Central Orchestrator Worker
- Runs the `OrchestrationWorkflow` that coordinates execution across servers
- Listens on the `execution-orchestrator` task queue
- Manages rollout strategies (Parallel, Sequential, Rolling)
- Tracks overall execution progress and handles failures

### Local Workers
- One worker runs on each target server
- Each worker listens on a server-specific task queue (e.g., `server-1`, `server-2`)
- Executes the `ServerExecutionWorkflow` with steps specific to that server
- Handles step execution, retries, and rollbacks

## Features

- **Multiple Rollout Strategies**
  - **Parallel**: Execute on all servers simultaneously
  - **Sequential**: Execute one server at a time
  - **Rolling**: Execute in batches with configurable batch size and delays

- **Step Execution Framework**
  - Extensible step handler system
  - Built-in handlers: `echo`, `script`, `sleep`, `file_write`
  - Easy to add custom step types

- **Error Handling**
  - Configurable retry policies
  - Continue-on-failure support for non-critical steps
  - Automatic rollback on required step failures
  - Max failures threshold to stop rollouts early

## Project Structure

```
kitsune/
├── cmd/
│   ├── local-worker/          # Worker that runs on each server
│   └── orchestration-worker/  # Central orchestration coordinator
├── pkg/
│   ├── activities/            # Activity implementations
│   │   ├── handlers/          # Step handler implementations
│   │   ├── step_activities.go
│   │   └── step_handler.go
│   ├── models/                # Data models and types
│   │   └── types.go
│   └── workflows/             # Workflow implementations
│       ├── execution.go       # Server-level workflow
│       └── orchestration.go   # Orchestration workflow
└── dev/                       # Development utilities
    ├── docker-compose.yaml
    ├── test-orchestration.sh
    └── test-plan-orchestrator.json
```

## Prerequisites

- Go 1.25.3+
- Docker and Docker Compose (for local development)
- Temporal Server

## Getting Started

### Local Development with Docker Compose

The project includes a complete Docker Compose setup with Temporal, PostgreSQL, and mock servers:

```bash
cd dev
docker-compose up -d
```

This starts:
- PostgreSQL (port 5432)
- Temporal Server (port 7233)
- Temporal UI (port 8080)
- Central Orchestrator Worker
- 3 Mock Server Workers (`server-1`, `server-2`, `server-3`)

### Running the Test Orchestration

```bash
cd dev
./test-orchestration.sh
```

This script:
1. Starts all services
2. Checks health of workers
3. Triggers an orchestration workflow across 3 servers
4. Shows execution results
5. Provides links to view in Temporal UI

### Manual Execution

#### 1. Start the Central Orchestrator Worker

```bash
export TEMPORAL_ADDRESS=localhost:7233
go run cmd/orchestration-worker/main.go
```

#### 2. Start Local Workers on Each Server

On each target server:

```bash
export SERVER_ID=<server-name>
export TEMPORAL_ADDRESS=<temporal-host>:7233
go run cmd/local-worker/main.go
```

#### 3. Trigger an Orchestration

Using the Temporal CLI:

```bash
temporal workflow start \
  --task-queue execution-orchestrator \
  --type OrchestrationWorkflow \
  --input '{
    "servers": ["server-1", "server-2", "server-3"],
    "steps": [
      {
        "name": "deploy",
        "type": "script",
        "params": {
          "command": "deploy.sh",
          "args": ["--version", "v1.2.3"]
        },
        "required": true
      }
    ],
    "rolloutStrategy": {
      "type": "Rolling",
      "batchSize": 1,
      "batchDelaySeconds": 30,
      "maxFailures": 1
    }
  }'
```

## Configuration

### Execution Request

```json
{
  "servers": ["server-1", "server-2"],
  "steps": [
    {
      "name": "step-name",
      "type": "echo|script|sleep|file_write",
      "params": {},
      "required": true,
      "continueOnFailure": false
    }
  ],
  "rolloutStrategy": {
    "type": "Parallel|Sequential|Rolling",
    "batchSize": 1,
    "batchDelaySeconds": 0,
    "maxFailures": 0,
    "canaryPercentage": 10
  }
}
```

### Step Types

#### Echo
Simple logging step:
```json
{
  "name": "log-message",
  "type": "echo",
  "params": {
    "message": "Starting deployment"
  }
}
```

#### Script
Execute shell scripts:
```json
{
  "name": "run-deploy",
  "type": "script",
  "params": {
    "command": "deploy.sh",
    "args": ["--version", "v1.2.3"]
  }
}
```

#### Sleep
Add delays:
```json
{
  "name": "wait",
  "type": "sleep",
  "params": {
    "duration": "30s"
  }
}
```

#### File Write
Write files to disk:
```json
{
  "name": "write-config",
  "type": "file_write",
  "params": {
    "path": "/etc/app/config.json",
    "content": "{\"key\": \"value\"}"
  }
}
```

### Rollout Strategies

#### Parallel
Execute on all servers at once:
```json
{
  "type": "Parallel"
}
```

#### Sequential
Execute one server at a time:
```json
{
  "type": "Sequential",
  "maxFailures": 1
}
```

#### Rolling
Execute in batches:
```json
{
  "type": "Rolling",
  "batchSize": 2,
  "batchDelaySeconds": 60,
  "maxFailures": 2
}
```

## Adding Custom Step Handlers

1. Create a new handler in `pkg/activities/handlers/`:

```go
package handlers

import (
    "context"
    "go.temporal.io/sdk/activity"
)

type CustomHandler struct{}

func (h *CustomHandler) Execute(ctx context.Context, params map[string]interface{}) error {
    logger := activity.GetLogger(ctx)
    // Your implementation here
    return nil
}

func (h *CustomHandler) Rollback(ctx context.Context, params map[string]interface{}) error {
    logger := activity.GetLogger(ctx)
    // Rollback logic here
    return nil
}
```

2. Register it in `cmd/local-worker/main.go`:

```go
registry.Register("custom", &handlers.CustomHandler{})
```

## Monitoring

### Temporal UI

Access the Temporal UI at http://localhost:8080 to:
- View workflow execution history
- See step-by-step progress
- Debug failures and retries
- Inspect workflow inputs and outputs

### Workflow Status

Check workflow status via CLI:

```bash
temporal workflow describe --workflow-id <workflow-id>
```

### View Logs

For Docker Compose setup:

```bash
docker logs kitsune-orchestrator
docker logs kitsune-mock-server-1
docker logs kitsune-mock-server-2
docker logs kitsune-mock-server-3
```

## Error Handling

### Required Steps
If a step is marked as `required: true` and fails:
1. Workflow execution stops
2. Automatic rollback is triggered for already-executed steps
3. Workflow returns an error

### Non-Required Steps
If a step has `continueOnFailure: true`:
1. Failure is logged but execution continues
2. Overall workflow can still succeed

### Max Failures
Configure `maxFailures` in rollout strategy:
- `0`: Stop on first failure
- `N`: Allow up to N server failures before stopping rollout

## Development

### Build

```bash
go build ./cmd/local-worker
go build ./cmd/orchestration-worker
```

### Run Tests

```bash
go test ./...
```

## License

MIT

## Contributing

Contributions welcome! Please open an issue or submit a pull request.
