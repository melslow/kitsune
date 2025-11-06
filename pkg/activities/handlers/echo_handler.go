package handlers

import (
	"context"
	"fmt"
	
	"go.temporal.io/sdk/activity"

	"github.com/melslow/kitsune/pkg/activities"
	"github.com/melslow/kitsune/pkg/activities/params"
)

type EchoParams struct {
	Message string `json:"message" validate:"required"`
}

// EchoHandler is the simplest handler - just logs a message
type EchoHandler struct{}

func (h *EchoHandler) Execute(ctx context.Context, rawParams map[string]interface{}) (activities.ExecutionMetadata, error) {
	var p EchoParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return nil, err
	}
	
	logger := activity.GetLogger(ctx)
	logger.Info("Echo", "message", p.Message)
	fmt.Printf("ECHO: %s\n", p.Message)
	
	return nil, nil
}

func (h *EchoHandler) Rollback(ctx context.Context, rawParams map[string]interface{}, metadata activities.ExecutionMetadata) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Echo rollback - nothing to do")
	return nil
}