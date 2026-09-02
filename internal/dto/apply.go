// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/arduino/go-paths-helper"
)

var fdtCmdName = "fdtoverlay"

type DeviceTreeApplier interface {
	// Builds and optionally applies device tree overlay commands.
	// Returns the command string used to configure the device tree.
	Apply(ctx context.Context, overlays []string, dryRun bool) (string, error)
}

type UnoQ struct {
	BaseDtbFile string
	OverlaysDir *paths.Path
	DtbFileName string
}

type VentunoQ struct {
	BaseDtbFile string
	OverlaysDir *paths.Path
	DtbFileName string
}

func (b UnoQ) Apply(ctx context.Context, overlays []string, dryRun bool) (string, error) {
	slices.Sort(overlays)
	overlays = slices.Compact(overlays)

	// Generate unique temp file name using nanosecond timestamp to prevent
	// race conditions when multiple instances run concurrently
	tempFileName := fmt.Sprintf("temporaryDeviceTree.%d.temp", time.Now().UnixNano())
	temporaryDtb := b.OverlaysDir.Join(tempFileName)

	args := buildOverlayCommand(b.OverlaysDir, b.BaseDtbFile, temporaryDtb, overlays)
	command := strings.Join(args, " ")

	if dryRun {
		return command, nil
	}

	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return command, fmt.Errorf("failed to create process: %w", err)
	}

	_, stderr, err := cmd.RunAndCaptureOutput(ctx)
	if err != nil {
		return command, fmt.Errorf("fdtoverlay failed with command %v: %w (stderr: %s)", args, err, stderr)
	}

	defer func() { _ = temporaryDtb.Remove() }()
	var destinationDtb = b.OverlaysDir.Join(b.DtbFileName)
	return command, moveDeviceTree(temporaryDtb, destinationDtb)
}

func (b VentunoQ) Apply(ctx context.Context, overlays []string, dryRun bool) (string, error) {
	// mount the device tree partition dtb_a
	mountPoint, err := os.MkdirTemp("/tmp", "dtb_")
	if err != nil {
		return "", fmt.Errorf("failed to create mountPoint: %w", err)
	}

	dtbA := "/dev/disk/by-partlabel/dtb_a"
	deferFunc, mountCmd, err := mountDeviceTree(dtbA, mountPoint, dryRun)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(mountPoint)
	defer deferFunc()

	slices.Sort(overlays)
	overlays = slices.Compact(overlays)

	// Generate unique temp file name using nanosecond timestamp to prevent
	// race conditions when multiple instances run concurrently
	tempFileName := fmt.Sprintf("temporaryDeviceTree.%d.temp", time.Now().UnixNano())
	temporaryDtb := paths.New(mountPoint).Join(tempFileName)

	args := buildOverlayCommand(b.OverlaysDir, b.BaseDtbFile, temporaryDtb, overlays)
	command := strings.Join(args, " ")

	if dryRun {
		return fmt.Sprintf("%s\n%s", mountCmd, command), nil
	}

	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return command, fmt.Errorf("failed to create process: %w", err)
	}

	_, stderr, err := cmd.RunAndCaptureOutput(ctx)
	if err != nil {
		return command, fmt.Errorf("fdtoverlay failed with command %v: %w (stderr: %s)", args, err, stderr)
	}

	defer func() { _ = temporaryDtb.Remove() }()
	var destinationDtb = paths.New(mountPoint).Join(b.DtbFileName)
	return command, moveDeviceTree(temporaryDtb, destinationDtb)
}
