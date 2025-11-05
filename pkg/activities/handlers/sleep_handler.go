package handlers

import (
	"context"
	"fmt"
	"time"
	
	"go.temporal.io/sdk/activity"
)

type SleepHandler struct{}

func (h *SleepHandler) Execute(ctx context.Context, params map[string]interface{}) error {
	logger := activity.GetLogger(ctx)
	
	durationSec, ok := params["duration"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid 'duration' parameter")
	}
	
	duration := time.Duration(durationSec) * time.Second
	
	logger.Info("Sleeping", "duration", duration)
	time.Sleep(duration)
	logger.Info("Sleep completed")
	
	return nil
}

func (h *SleepHandler) Rollback(ctx context.Context, params map[string]interface{}) error {
	return nil
}