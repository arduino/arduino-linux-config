// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/sync"
)

func mountDeviceTree(source, mountPoint string, dryRun bool) (func(), string, error) {
	mountArgs := []string{"mount", "-t", "vfat", source, mountPoint}
	dryRunStr := strings.Join(mountArgs, " ")

	if dryRun {
		return func() {}, dryRunStr, nil
	}

	cmd, err := paths.NewProcess(nil, mountArgs...)
	if err != nil {
		return nil, dryRunStr, fmt.Errorf("failed to create mount process: %w", err)
	}

	if _, stderr, err := cmd.RunAndCaptureOutput(context.Background()); err != nil {
		return nil, dryRunStr, fmt.Errorf("failed to mount %s to %s: %w: %s", source, mountPoint, err, stderr)
	}

	return func() {
		sync.SyncToDisk()

		unmountArgs := []string{"umount", "-l", mountPoint}
		cmd, err := paths.NewProcess(nil, unmountArgs...)
		if err != nil {
			slog.Error("failed to create unmount process", "mountPoint", mountPoint, "error", err)
			return
		}

		if _, stderr, unmountErr := cmd.RunAndCaptureOutput(context.Background()); unmountErr != nil {
			slog.Error("failed to unmount", "mountPoint", mountPoint, "error", unmountErr, "output", stderr)
		}
	}, dryRunStr, nil
}

func buildOverlayCommand(overlaysDir *paths.Path, baseDtbFile string, temporaryDtb *paths.Path, overlays []string) []string {
	var baseDTB = overlaysDir.Join(baseDtbFile)
	if len(overlays) == 0 {
		return []string{"cp", baseDTB.String(), temporaryDtb.String()}
	}

	overlayFullPaths := make([]string, len(overlays))
	for i, overlay := range overlays {
		overlayFullPaths[i] = overlaysDir.Join(overlay).String()
	}

	args := append([]string{fdtBinary.String(), "-i", baseDTB.String(), "-o", temporaryDtb.String()}, overlayFullPaths...)

	return args
}

func moveDeviceTree(temporaryDtb *paths.Path, destinationDtb *paths.Path) error {
	if err := temporaryDtb.Rename(destinationDtb); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", temporaryDtb, destinationDtb, err)
	}

	// Flush kernel buffers to disk to ensure the DTB is persisted
	// before the system potentially reboots or loses power.
	sync.SyncToDisk()
	return nil
}
