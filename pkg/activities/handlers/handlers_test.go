package handlers

import (
	"context"
	"strings"
	"testing"
)

func TestEchoHandler_RejectsUnsupportedParams(t *testing.T) {
	h := &EchoHandler{}
	ctx := context.Background()

	params := map[string]interface{}{
		"message":     "test",
		"unsupported": "param",
	}

	_, err := h.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for unsupported parameter")
	}

	if !strings.Contains(err.Error(), "unsupported parameters: unsupported") {
		t.Errorf("Expected unsupported parameters error, got: %v", err)
	}
}

func TestEchoHandler_RequiresMessage(t *testing.T) {
	h := &EchoHandler{}
	ctx := context.Background()

	params := map[string]interface{}{}

	_, err := h.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for missing message parameter")
	}

	if !strings.Contains(err.Error(), "missing required parameter: message") {
		t.Errorf("Expected missing required parameter error, got: %v", err)
	}
}

func TestSleepHandler_RejectsUnsupportedParams(t *testing.T) {
	h := &SleepHandler{}
	ctx := context.Background()

	params := map[string]interface{}{
		"duration":    1.0,
		"unsupported": "param",
	}

	_, err := h.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for unsupported parameter")
	}

	if !strings.Contains(err.Error(), "unsupported parameters: unsupported") {
		t.Errorf("Expected unsupported parameters error, got: %v", err)
	}
}

func TestFileWriteHandler_RejectsUnsupportedParams(t *testing.T) {
	h := &FileWriteHandler{}
	ctx := context.Background()

	params := map[string]interface{}{
		"path":        "/tmp/test",
		"content":     "test",
		"unsupported": "param",
	}

	_, err := h.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for unsupported parameter")
	}

	if !strings.Contains(err.Error(), "unsupported parameters: unsupported") {
		t.Errorf("Expected unsupported parameters error, got: %v", err)
	}
}

func TestScriptHandler_RejectsUnsupportedParams(t *testing.T) {
	h := &ScriptHandler{}
	ctx := context.Background()

	params := map[string]interface{}{
		"script":      "/bin/echo",
		"unsupported": "param",
	}

	_, err := h.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for unsupported parameter")
	}

	if !strings.Contains(err.Error(), "unsupported parameters: unsupported") {
		t.Errorf("Expected unsupported parameters error, got: %v", err)
	}
}

func TestYumUpgradeHandler_RejectsUnsupportedParams(t *testing.T) {
	h := &YumUpgradeHandler{}
	ctx := context.Background()

	params := map[string]interface{}{
		"package":     "nginx",
		"version":     "1.20.0",
		"unsupported": "param",
	}

	_, err := h.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for unsupported parameter")
	}

	if !strings.Contains(err.Error(), "unsupported parameters: unsupported") {
		t.Errorf("Expected unsupported parameters error, got: %v", err)
	}
}
