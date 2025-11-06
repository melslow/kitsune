package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.temporal.io/sdk/activity"

	"github.com/melslow/kitsune/pkg/activities"
	"github.com/melslow/kitsune/pkg/activities/params"
)

type YumUpgradeParams struct {
	Package string `json:"package" validate:"required"`
	Version string `json:"version" validate:"required"`
}

type YumUpgradeHandler struct{}

func (h *YumUpgradeHandler) Execute(ctx context.Context, rawParams map[string]interface{}) (activities.ExecutionMetadata, error) {
	var p YumUpgradeParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return nil, err
	}

	logger := activity.GetLogger(ctx)
	logger.Info("Starting yum upgrade", "package", p.Package, "version", p.Version)

	metadata := make(activities.ExecutionMetadata)

	// Get currently installed version for rollback
	cmd := exec.CommandContext(ctx, "rpm", "-q", p.Package, "--queryformat", "%{VERSION}-%{RELEASE}")
	output, err := cmd.CombinedOutput()
	if err == nil {
		previousVersion := strings.TrimSpace(string(output))
		metadata["previous_version"] = previousVersion
		logger.Info("Captured current version for rollback", "package", p.Package, "currentVersion", previousVersion)
	} else {
		logger.Warn("Could not get current version", "package", p.Package, "error", err.Error())
	}

	// Perform the upgrade
	fullPackage := fmt.Sprintf("%s-%s", p.Package, p.Version)
	cmd = exec.CommandContext(ctx, "yum", "upgrade", "-y", fullPackage)
	output, err = cmd.CombinedOutput()

	logger.Info("Yum upgrade completed", "output", string(output))

	if err != nil {
		return metadata, fmt.Errorf("yum upgrade failed: %w, output: %s", err, string(output))
	}

	return metadata, nil
}

func (h *YumUpgradeHandler) Rollback(ctx context.Context, rawParams map[string]interface{}, metadata activities.ExecutionMetadata) error {
	var p YumUpgradeParams
	if err := params.ParseAndValidate(rawParams, &p); err != nil {
		return err
	}

	logger := activity.GetLogger(ctx)
	previousVersion, ok := metadata["previous_version"].(string)
	if !ok || previousVersion == "" {
		logger.Warn("No previous version captured, cannot rollback", "package", p.Package)
		return fmt.Errorf("no previous version available for rollback")
	}

	logger.Info("Starting rollback for package", "package", p.Package, "targetVersion", previousVersion)

	// Check current installed version to make intelligent rollback decision
	cmd := exec.CommandContext(ctx, "rpm", "-q", p.Package, "--queryformat", "%{VERSION}-%{RELEASE}")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Package might not be installed or in inconsistent state
		logger.Warn("Could not query current package version, attempting rollback anyway", "package", p.Package, "error", err.Error())
	} else {
		currentVersion := strings.TrimSpace(string(output))
		logger.Info("Current package version", "package", p.Package, "version", currentVersion)

		// If current version is already the previous version, no rollback needed
		if currentVersion == previousVersion {
			logger.Info("Package already at target version, no rollback needed", "package", p.Package, "version", currentVersion)
			return nil
		}

		// Log what we're doing
		logger.Info("Package version differs from target, proceeding with rollback",
			"package", p.Package,
			"currentVersion", currentVersion,
			"targetVersion", previousVersion)
	}

	// Downgrade to previous version
	fullPackage := fmt.Sprintf("%s-%s", p.Package, previousVersion)
	cmd = exec.CommandContext(ctx, "yum", "downgrade", "-y", fullPackage)
	output, err = cmd.CombinedOutput()

	logger.Info("Yum downgrade completed", "output", string(output))

	if err != nil {
		return fmt.Errorf("yum downgrade failed: %w, output: %s", err, string(output))
	}

	// Verify rollback was successful
	cmd = exec.CommandContext(ctx, "rpm", "-q", p.Package, "--queryformat", "%{VERSION}-%{RELEASE}")
	output, err = cmd.CombinedOutput()
	if err == nil {
		finalVersion := strings.TrimSpace(string(output))
		if finalVersion == previousVersion {
			logger.Info("Rollback verified successful", "package", p.Package, "version", finalVersion)
		} else {
			logger.Warn("Rollback may not have succeeded", "package", p.Package, "expectedVersion", previousVersion, "actualVersion", finalVersion)
		}
	}

	return nil
}
