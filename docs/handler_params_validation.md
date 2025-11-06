# Handler Parameter Validation

All handlers now use predefined input structs with validation to ensure only supported parameters are accepted.

## Validation Levels

### 1. Workflow-Level Validation (Early Detection)
Before dispatching tasks to workers, the orchestration and execution workflows validate all step parameters. This catches errors early and prevents wasting resources on invalid configurations.

**Benefits:**
- Fail fast before any execution starts
- No resources wasted on workers
- Consistent error messages across all steps
- Validates all steps in a plan at once

**Where it happens:**
- `OrchestrationWorkflow` - validates before distributing to servers
- `ServerExecutionWorkflow` - validates before executing steps on a server

### 2. Handler-Level Validation (Runtime Safety)
Each handler validates its parameters during execution as a safety net.

**Benefits:**
- Defense-in-depth approach
- Protects against direct activity calls
- Ensures type safety at runtime

## How It Works

Each handler defines a typed parameter struct with:
- `json` tags to specify parameter names
- `validate:"required"` tags for required fields
- `omitempty` for optional fields

The `params.ParseAndValidate()` function:
1. Checks for unsupported parameters and rejects them with an error
2. Parses parameters into the typed struct
3. Validates that all required fields are present

The `StepValidator` provides workflow-level validation:
1. `ValidateStep()` - validates a single step definition
2. `ValidateSteps()` - validates an entire list of steps
3. Returns detailed error messages indicating which step failed

## Handler Parameters

### EchoHandler
```go
type EchoParams struct {
    Message string `json:"message" validate:"required"`
}
```

### SleepHandler
```go
type SleepParams struct {
    Duration float64 `json:"duration" validate:"required"`
}
```

### FileWriteHandler
```go
type FileWriteParams struct {
    Path    string `json:"path" validate:"required"`
    Content string `json:"content" validate:"required"`
}
```

### ScriptHandler
```go
type ScriptParams struct {
    Script         string   `json:"script" validate:"required"`
    Args           []string `json:"args,omitempty"`
    RollbackScript string   `json:"rollback_script,omitempty"`
}
```

### YumUpgradeHandler
```go
type YumUpgradeParams struct {
    Package string `json:"package" validate:"required"`
    Version string `json:"version" validate:"required"`
}
```

## Error Messages

### Missing Required Parameter
```
missing required parameter: message
```

### Unsupported Parameter
```
unsupported parameters: typo, invalid_param
```

## Examples

### Valid Request
```json
{
  "type": "echo",
  "params": {
    "message": "Hello, World!"
  }
}
```

### Invalid - Unsupported Parameter
```json
{
  "type": "echo",
  "params": {
    "message": "Hello",
    "unsupported": "value"
  }
}
```
**Workflow-level error:** `step validation failed: step 1: validation failed for step 'test echo' (type: echo): unsupported parameters: unsupported`

**Handler-level error:** `unsupported parameters: unsupported`

### Invalid - Missing Required Parameter
```json
{
  "type": "yum_upgrade",
  "params": {
    "package": "nginx"
  }
}
```
**Workflow-level error:** `step validation failed: step 1: validation failed for step 'upgrade nginx' (type: yum_upgrade): missing required parameter: version`

**Handler-level error:** `missing required parameter: version`

### Multiple Steps Validation
When validating multiple steps, the validator will identify which step has the error:

```go
steps := []StepDefinition{
  {Name: "step 1", Type: "echo", Params: map[string]interface{}{"message": "ok"}},
  {Name: "step 2", Type: "sleep", Params: map[string]interface{}{"duration": 1.0, "typo": "bad"}},
  {Name: "step 3", Type: "yum_upgrade", Params: map[string]interface{}{"package": "nginx", "version": "1.20"}},
}
```

Error: `step 2: validation failed for step 'step 2' (type: sleep): unsupported parameters: typo`

This ensures that configuration errors are caught before any execution begins.
