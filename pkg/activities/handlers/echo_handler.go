package handlers

import (
	"context"
	"fmt"
	
	"go.temporal.io/sdk/activity"
)

// EchoHandler is the simplest handler - just logs a message
type EchoHandler struct{}

func (h *EchoHandler) Execute(ctx context.Context, params map[string]interface{}) error {
	logger := activity.GetLogger(ctx)
	
	message, ok := params["message"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'message' parameter")
	}
	
	logger.Info("Echo", "message", message)
	fmt.Printf("ECHO: %s\n", message)
	
	return nil
}

func (h *EchoHandler) Rollback(ctx context.Context, params map[string]interface{}) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Echo rollback - nothing to do")
	return nil
}