package handlers

import (
	"context"
	"fmt"
	"time"
	
	"go.temporal.io/sdk/activity"

	"github.com/melslow/kitsune/pkg/activities"
)

type SleepHandler struct{}

func (h *SleepHandler) Execute(ctx context.Context, params map[string]interface{}) (activities.ExecutionMetadata, error) {
	logger := activity.GetLogger(ctx)
	
	durationSec, ok := params["duration"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'duration' parameter")
	}
	
	duration := time.Duration(durationSec) * time.Second
	
	logger.Info("Sleeping", "duration", duration)
	time.Sleep(duration)
	logger.Info("Sleep completed")
	
	return nil, nil
}

func (h *SleepHandler) Rollback(ctx context.Context, params map[string]interface{}, metadata activities.ExecutionMetadata) error {
	return nil
}