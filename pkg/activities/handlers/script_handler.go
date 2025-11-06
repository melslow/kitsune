package handlers

import (
	"context"
	"fmt"
	"os/exec"
	
	"go.temporal.io/sdk/activity"

	"github.com/melslow/kitsune/pkg/activities"
	"github.com/melslow/kitsune/pkg/activities/params"
)

type ScriptParams struct {
	Script         string   `json:"script" validate:"required"`
	Args           []string `json:"args,omitempty"`
	RollbackScript string   `json:"rollback_script,omitempty"`
}

type ScriptHandler struct{}

func (h *ScriptHandler) Execute(ctx context.Context, rawParams map[string]interface{}) (activities.ExecutionMetadata, error) {
	var p ScriptParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return nil, err
	}
	
	logger := activity.GetLogger(ctx)
	logger.Info("Running script", "script", p.Script)
	
	cmd := exec.CommandContext(ctx, p.Script, p.Args...)
	output, err := cmd.CombinedOutput()
	
	logger.Info("Script completed", "output", string(output))
	
	if err != nil {
		return nil, fmt.Errorf("script failed: %w, output: %s", err, string(output))
	}
	
	return nil, nil
}

func (h *ScriptHandler) Rollback(ctx context.Context, rawParams map[string]interface{}, metadata activities.ExecutionMetadata) error {
	var p ScriptParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return nil
	}
	
	logger := activity.GetLogger(ctx)
	if p.RollbackScript != "" {
		logger.Info("Running rollback script", "script", p.RollbackScript)
		cmd := exec.CommandContext(ctx, p.RollbackScript)
		return cmd.Run()
	}
	
	logger.Info("No rollback script specified")
	return nil
}