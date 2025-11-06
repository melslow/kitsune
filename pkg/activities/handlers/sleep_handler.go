package handlers

import (
	"context"
	"time"
	
	"go.temporal.io/sdk/activity"

	"github.com/melslow/kitsune/pkg/activities"
	"github.com/melslow/kitsune/pkg/activities/params"
)

type SleepParams struct {
	Duration float64 `json:"duration" validate:"required"`
}

type SleepHandler struct{}

func (h *SleepHandler) Execute(ctx context.Context, rawParams map[string]interface{}) (activities.ExecutionMetadata, error) {
	var p SleepParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return nil, err
	}
	
	logger := activity.GetLogger(ctx)
	duration := time.Duration(p.Duration) * time.Second
	
	logger.Info("Sleeping", "duration", duration)
	time.Sleep(duration)
	logger.Info("Sleep completed")
	
	return nil, nil
}

func (h *SleepHandler) Rollback(ctx context.Context, rawParams map[string]interface{}, metadata activities.ExecutionMetadata) error {
	return nil
}