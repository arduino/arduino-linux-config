// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/arduino/go-paths-helper"
)

var fdtoverlayPath = paths.New("/usr/bin/fdtoverlay")

var qcomDTBDir = paths.New("/boot/efi/dtb/qcom/")
var systemDTB = qcomDTBDir.Join("qrb2210-arduino-imola.dtb")
var baseDTB = qcomDTBDir.Join("qrb2210-arduino-imola-base.dtb")

func Apply(ctx context.Context, overlays []string) error {
	if len(overlays) == 0 {
		return nil
	}

	slices.Sort(overlays)
	overlays = slices.Compact(overlays)

	args, tempFile := buildFdtoverlayCommandAt(overlays, time.Now())

	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	defer func() { _ = tempFile.Remove() }()

	_, stderr, err := cmd.RunAndCaptureOutput(ctx)
	if err != nil {
		return fmt.Errorf("fdtoverlay failed with command %v: %w (stderr: %s)", args, err, stderr)
	}

	if err := tempFile.Rename(systemDTB); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", tempFile, systemDTB, err)
	}

	return nil
}

func buildFdtoverlayCommandAt(overlays []string, now time.Time) ([]string, *paths.Path) {
	// Generate unique temp file name using nanosecond timestamp to prevent
	// race conditions when multiple instances run concurrently
	tempFileName := fmt.Sprintf("qrb2210-arduino-imola.dtb.%d.temp", now.UnixNano())
	temporaryDtb := qcomDTBDir.Join(tempFileName)

	overlayPaths := make([]string, len(overlays))
	for i, overlay := range overlays {
		overlayPaths[i] = qcomDTBDir.Join(overlay).String()
	}

	args := append([]string{fdtoverlayPath.String(), "-i", baseDTB.String(), "-o", temporaryDtb.String()}, overlayPaths...)

	return args, temporaryDtb
}
