// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/sync"
)

var fdtBinary = paths.New("/usr/bin/fdtoverlay")

type Board interface {
	Apply(ctx context.Context, overlays []string) error
	buildOverlayCommand(overlaysDir *paths.Path, overlays []string, now time.Time) ([]string, *paths.Path)
	moveDeviceTree(temporaryDtb *paths.Path, destinationDtb *paths.Path) error
}

type UnoQ struct {
}

type VentunoQ struct {
}

func (b UnoQ) moveDeviceTree(temporaryDtb *paths.Path, destinationDtb *paths.Path) error {
	if err := temporaryDtb.Rename(destinationDtb); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", temporaryDtb, destinationDtb, err)
	}

	// Flush kernel buffers to disk to ensure the DTB is persisted
	// before the system potentially reboots or loses power.
	sync.SyncToDisk()
	return nil
}

func (b UnoQ) buildOverlayCommand(overlaysDir *paths.Path, overlays []string, now time.Time) ([]string, *paths.Path) {
	// Generate unique temp file name using nanosecond timestamp to prevent
	// race conditions when multiple instances run concurrently
	tempFileName := fmt.Sprintf("temporaryDeviceTree.%d.temp", now.UnixNano())
	temporaryDtb := overlaysDir.Join(tempFileName)

	overlayFullPaths := make([]string, len(overlays))
	for i, overlay := range overlays {
		overlayFullPaths[i] = overlaysDir.Join(overlay).String()
	}

	var baseDTB = overlaysDir.Join("qrb2210-arduino-imola-base.dtb")
	args := append([]string{fdtBinary.String(), "-i", baseDTB.String(), "-o", temporaryDtb.String()}, overlayFullPaths...)

	return args, temporaryDtb
}

func (b UnoQ) Apply(ctx context.Context, overlays []string) error {
	if len(overlays) == 0 {
		return nil
	}

	slices.Sort(overlays)
	overlays = slices.Compact(overlays)

	var overlaysDir = paths.New("/boot/efi/dtb/qcom/")
	args, tempFile := b.buildOverlayCommand(overlaysDir, overlays, time.Now())

	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	defer func() { _ = tempFile.Remove() }()

	_, stderr, err := cmd.RunAndCaptureOutput(ctx)
	if err != nil {
		return fmt.Errorf("fdtoverlay failed with command %v: %w (stderr: %s)", args, err, stderr)
	}

	var destinationDtb = overlaysDir.Join("qrb2210-arduino-imola.dtb")
	return b.moveDeviceTree(tempFile, destinationDtb)
}

func (b VentunoQ) buildOverlayCommand(overlaysDir *paths.Path, overlays []string, now time.Time) ([]string, *paths.Path) {
	slog.Error("buildOverlayCommand not yet implemented on this platform.")
	// This should be the same, to be refactored
	// Check the amount of available space in the destination partition
	return nil, nil
}

func (b VentunoQ) Apply(ctx context.Context, overlays []string) error {
	// mount
	// the same of unoQ apply
	// umount the dir after moving
	return fmt.Errorf("feature not yet implemented on this platform.")
}

func (b VentunoQ) moveDeviceTree(temporaryDtb *paths.Path, destinationDtb *paths.Path) error {
	slog.Error("moveDeviceTree not yet implemented on this platform.")
	// do the same of the move in the mounted dir
	return nil
}
