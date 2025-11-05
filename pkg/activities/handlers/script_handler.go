package handlers

import (
	"context"
	"fmt"
	"os/exec"
	
	"go.temporal.io/sdk/activity"
)

type ScriptHandler struct{}

func (h *ScriptHandler) Execute(ctx context.Context, params map[string]interface{}) error {
	logger := activity.GetLogger(ctx)
	
	script, ok := params["script"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'script' parameter")
	}
	
	logger.Info("Running script", "script", script)
	
	// Get args if present
	var args []string
	if argsParam, ok := params["args"].([]interface{}); ok {
		for _, arg := range argsParam {
			if argStr, ok := arg.(string); ok {
				args = append(args, argStr)
			}
		}
	}
	
	cmd := exec.CommandContext(ctx, script, args...)
	output, err := cmd.CombinedOutput()
	
	logger.Info("Script completed", "output", string(output))
	
	if err != nil {
		return fmt.Errorf("script failed: %w, output: %s", err, string(output))
	}
	
	return nil
}

func (h *ScriptHandler) Rollback(ctx context.Context, params map[string]interface{}) error {
	logger := activity.GetLogger(ctx)
	
	if rollbackScript, ok := params["rollback_script"].(string); ok {
		logger.Info("Running rollback script", "script", rollbackScript)
		cmd := exec.CommandContext(ctx, rollbackScript)
		return cmd.Run()
	}
	
	logger.Info("No rollback script specified")
	return nil
}