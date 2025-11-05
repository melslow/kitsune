package handlers

import (
	"context"
	"fmt"
	"os"
	
	"go.temporal.io/sdk/activity"
)

type FileWriteHandler struct{}

func (h *FileWriteHandler) Execute(ctx context.Context, params map[string]interface{}) error {
	logger := activity.GetLogger(ctx)
	
	path, ok := params["path"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'path' parameter")
	}
	
	content, ok := params["content"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'content' parameter")
	}
	
	logger.Info("Writing file", "path", path)
	
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	logger.Info("File written successfully")
	return nil
}

func (h *FileWriteHandler) Rollback(ctx context.Context, params map[string]interface{}) error {
	logger := activity.GetLogger(ctx)
	
	path, ok := params["path"].(string)
	if !ok {
		return nil
	}
	
	logger.Info("Deleting file for rollback", "path", path)
	return os.Remove(path)
}