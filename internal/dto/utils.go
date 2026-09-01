// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"

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

func buildOverlayCommand(overlaysDir *paths.Path, src string, dst *paths.Path, overlays []string) []string {
	if len(overlays) == 0 {
		return []string{"cp", src, dst.String()}
	}

	overlayFullPaths := make([]string, len(overlays))
	for i, overlay := range overlays {
		overlayFullPaths[i] = overlaysDir.Join(overlay).String()
	}

	args := append([]string{fdtCmdName, "-i", src, "-o", dst.String()}, overlayFullPaths...)

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

// On VentunoQ with Ubuntu the device tree is shipped as a combined dtb:
// it is unpacked, the overlays are applied, and it is packed back.
const (
	// Every flattened device tree starts with this magic, followed by its total size.
	fdtMagic              = 0xd00dfeed
	fdtHeaderSize         = 8
	fdtAlignment          = 8
	monzaCompatibleString = "arduino,monza"
)

// The device trees extracted from a combined dtb, in the order they are stored in it.
type unpackedDeviceTree struct {
	mountPoint  *paths.Path
	monza       *paths.Path
	monzaIndex  int
	deviceTrees [][]byte
}

// Splits a combined dtb into its device trees and writes the one describing the Monza
// board into the mount point, so that fdtoverlay can take it as input.
func unpack(exec executor.Executor, combinedDtb string, mountPoint string) (string, *unpackedDeviceTree, error) {
	data, err := os.ReadFile(combinedDtb)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read %s: %w", combinedDtb, err)
	}

	deviceTrees, err := splitDeviceTrees(data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to split %s: %w", combinedDtb, err)
	}

	unpacked := &unpackedDeviceTree{
		mountPoint:  paths.New(mountPoint),
		monzaIndex:  -1,
		deviceTrees: deviceTrees,
	}
	for i, deviceTree := range deviceTrees {
		if bytes.Contains(deviceTree, []byte(monzaCompatibleString)) {
			unpacked.monzaIndex = i
			break
		}
	}

	if unpacked.monzaIndex < 0 {
		return "", nil, fmt.Errorf("no %q device tree found in %s", monzaCompatibleString, combinedDtb)
	}

	unpacked.monza = unpacked.mountPoint.Join("monza.dtb")
	if err := exec.WriteFile(unpacked.monza, deviceTrees[unpacked.monzaIndex], 0600); err != nil {
		return "", nil, fmt.Errorf("failed to write %s: %w", unpacked.monza, err)
	}

	return unpacked.monza.String(), unpacked, nil
}

// Puts the customized device tree back in place of the Monza one and rebuilds the
// combined dtb, returning the path of the temporary file that holds it.
func pack(exec executor.Executor, temporaryDtb *paths.Path, unpacked *unpackedDeviceTree) (*paths.Path, error) {
	customized, err := os.ReadFile(temporaryDtb.String())
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read %s: %w", temporaryDtb, err)
		}
		// Dry run: no device tree was actually customized, so the original one is used.
		customized = unpacked.deviceTrees[unpacked.monzaIndex]
	}
	unpacked.deviceTrees[unpacked.monzaIndex] = pad(customized)

	packedDtb := unpacked.mountPoint.Join(temporaryDtbName())
	if err := exec.WriteFile(packedDtb, bytes.Join(unpacked.deviceTrees, nil), 0600); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", packedDtb, err)
	}

	return packedDtb, nil
}

// Each device tree keeps the padding that follows it, so that the untouched ones
// are written back byte by byte.
func splitDeviceTrees(data []byte) ([][]byte, error) {
	var deviceTrees [][]byte
	for offset := 0; offset+fdtHeaderSize <= len(data); {
		if !hasMagic(data, offset) {
			return nil, fmt.Errorf("missing device tree magic at offset %d", offset)
		}

		size := int(binary.BigEndian.Uint32(data[offset+4:]))
		if size < fdtHeaderSize || offset+size > len(data) {
			return nil, fmt.Errorf("invalid device tree size %d at offset %d", size, offset)
		}

		end := offset + size
		for end+fdtHeaderSize <= len(data) && !hasMagic(data, end) {
			end++
		}
		if end+fdtHeaderSize > len(data) {
			end = len(data)
		}

		deviceTrees = append(deviceTrees, data[offset:end])
		offset = end
	}

	if len(deviceTrees) == 0 {
		return nil, fmt.Errorf("no device tree found")
	}
	return deviceTrees, nil
}

func hasMagic(data []byte, offset int) bool {
	return binary.BigEndian.Uint32(data[offset:]) == fdtMagic
}

func pad(deviceTree []byte) []byte {
	remainder := len(deviceTree) % fdtAlignment
	if remainder == 0 {
		return deviceTree
	}
	return append(deviceTree, make([]byte, fdtAlignment-remainder)...)
}
