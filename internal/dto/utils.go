// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/executor"
)

// Returns the function that unmounts the partition.
func mountDeviceTree(ctx context.Context, exec executor.Executor, source, mountPoint string) (func(), error) {
	if err := exec.Run(ctx, "mount", "-t", "vfat", source, mountPoint); err != nil {
		return nil, fmt.Errorf("failed to mount %s to %s: %w", source, mountPoint, err)
	}

	return func() {
		exec.Sync()

		// Not the caller context: the unmount must run even after a cancellation.
		if err := exec.Run(context.Background(), "umount", "-l", mountPoint); err != nil {
			slog.Error("failed to unmount", "mountPoint", mountPoint, "error", err)
		}
	}, nil
}

var fdtCmdName = "fdtoverlay"

func buildOverlayCommand(overlaysDir *paths.Path, baseDtbFile string, temporaryDtb *paths.Path, overlays []string) []string {
	if len(overlays) == 0 {
		return []string{"cp", baseDtbFile, temporaryDtb.String()}
	}

	overlayFullPaths := make([]string, len(overlays))
	for i, overlay := range overlays {
		overlayFullPaths[i] = overlaysDir.Join(overlay).String()
	}

	args := append([]string{fdtCmdName, "-i", baseDtbFile, "-o", temporaryDtb.String()}, overlayFullPaths...)

	return args
}

func moveDeviceTree(exec executor.Executor, temporaryDtb *paths.Path, destinationDtb *paths.Path) error {
	if err := exec.Rename(temporaryDtb, destinationDtb); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", temporaryDtb, destinationDtb, err)
	}

	// Flush kernel buffers to disk to ensure the DTB is persisted
	// before the system potentially reboots or loses power.
	exec.Sync()
	return nil
}
