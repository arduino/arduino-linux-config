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

var fdtBinary = paths.New("/usr/bin/fdtoverlay")

type Board interface {
	Apply(ctx context.Context, overlays []string, dryRun bool) (string, error)
}

type UnoQ struct {
}

type VentunoQ struct {
}

func (b UnoQ) Apply(ctx context.Context, overlays []string, dryRun bool) (string, error) {
	slices.Sort(overlays)
	overlays = slices.Compact(overlays)

	var overlaysDir = paths.New("/boot/efi/dtb/qcom/")
	var baseDtbFile = "qrb2210-arduino-imola-base.dtb"

	// Generate unique temp file name using nanosecond timestamp to prevent
	// race conditions when multiple instances run concurrently
	tempFileName := fmt.Sprintf("temporaryDeviceTree.%d.temp", time.Now().UnixNano())
	temporaryDtb := overlaysDir.Join(tempFileName)

	args := buildOverlayCommand(overlaysDir, baseDtbFile, temporaryDtb, overlays)
	command := strings.Join(args, " ")

	if dryRun {
		return command, nil
	}

	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return command, fmt.Errorf("failed to create process: %w", err)
	}

	defer func() { _ = temporaryDtb.Remove() }()

	_, stderr, err := cmd.RunAndCaptureOutput(ctx)
	if err != nil {
		return command, fmt.Errorf("fdtoverlay failed with command %v: %w (stderr: %s)", args, err, stderr)
	}

	var destinationDtb = overlaysDir.Join("qrb2210-arduino-imola.dtb")
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

	var overlaysDir = paths.New("/var/lib/arduino-linux-config/overlays/ventunoq/jmedia-carrier/")
	var baseDtbFile = "qualcomm_technologies_inc._monaco_monza_addons.bin"

	// Generate unique temp file name using nanosecond timestamp to prevent
	// race conditions when multiple instances run concurrently
	tempFileName := fmt.Sprintf("temporaryDeviceTree.%d.temp", time.Now().UnixNano())
	temporaryDtb := paths.New(mountPoint).Join(tempFileName)

	args := buildOverlayCommand(overlaysDir, baseDtbFile, temporaryDtb, overlays)
	command := strings.Join(args, " ")

	if dryRun {
		return fmt.Sprintf("%s\n%s", mountCmd, command), nil
	}

	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return command, fmt.Errorf("failed to create process: %w", err)
	}

	defer func() { _ = temporaryDtb.Remove() }()

	_, stderr, err := cmd.RunAndCaptureOutput(ctx)
	if err != nil {
		return command, fmt.Errorf("fdtoverlay failed with command %v: %w (stderr: %s)", args, err, stderr)
	}

	var destinationDtb = paths.New(mountPoint).Join("combined-dtb.dtb")
	return command, moveDeviceTree(temporaryDtb, destinationDtb)
}
