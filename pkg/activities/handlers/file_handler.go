package handlers

import (
	"context"
	"fmt"
	"os"
	
	"go.temporal.io/sdk/activity"

	"github.com/melslow/kitsune/pkg/activities"
	"github.com/melslow/kitsune/pkg/activities/params"
)

type FileWriteParams struct {
	Path    string `json:"path" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type FileWriteHandler struct{}

func (h *FileWriteHandler) Execute(ctx context.Context, rawParams map[string]interface{}) (activities.ExecutionMetadata, error) {
	var p FileWriteParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return nil, err
	}
	
	logger := activity.GetLogger(ctx)
	logger.Info("Writing file", "path", p.Path)
	
	err := os.WriteFile(p.Path, []byte(p.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	
	logger.Info("File written successfully")
	return nil, nil
}

func (h *FileWriteHandler) Rollback(ctx context.Context, rawParams map[string]interface{}, metadata activities.ExecutionMetadata) error {
	var p FileWriteParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return nil
	}
	
	logger := activity.GetLogger(ctx)
	logger.Info("Deleting file for rollback", "path", p.Path)
	return os.Remove(p.Path)
}