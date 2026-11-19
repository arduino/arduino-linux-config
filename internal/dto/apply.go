// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package dto

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/executor"
)

type DeviceTreeApplier interface {
	// Applies the device tree overlays through the given executor.
	Apply(ctx context.Context, exec executor.Executor, overlays []string) error
}

type UnoQ struct {
	BaseDtbFile string
	OverlaysDir *paths.Path
	DtbFileName string
}

type VentunoQ struct {
	BaseDtbFullPath string
	OverlaysDir     *paths.Path
	DtbFileName     string
}

// VentunoQAddon plugs device-tree addons into the connectors of a base dtb.
// Each overlay is "<addon.dtba>@<connector path>".
type VentunoQAddon struct {
	AddonDir    *paths.Path // fdtaddon, dtc, the base dtb and the addons
	BaseDtbFile string
	DtbFileName string
}

func (b UnoQ) Apply(ctx context.Context, exec executor.Executor, overlays []string) error {
	temporaryDtb := b.OverlaysDir.Join(temporaryDtbName())
	defer func() { _ = exec.Remove(temporaryDtb) }()

	baseDtb := b.OverlaysDir.Join(b.BaseDtbFile)
	args := buildOverlayCommand(b.OverlaysDir, baseDtb.String(), temporaryDtb, uniqueOverlays(overlays))
	if err := exec.Run(ctx, args...); err != nil {
		return err
	}

	return moveDeviceTree(exec, temporaryDtb, b.OverlaysDir.Join(b.DtbFileName))
}

func (b VentunoQ) Apply(ctx context.Context, exec executor.Executor, overlays []string) error {
	mountPoint := paths.New("/run/arduino-linux-config/dtb")
	if err := exec.MkdirAll(mountPoint); err != nil {
		return fmt.Errorf("failed to create mountPoint: %w", err)
	}

	// mount the device tree partition dtb_a
	unmount, err := mountDeviceTree(ctx, exec, "/dev/disk/by-partlabel/dtb_a", mountPoint.String())
	if err != nil {
		return err
	}
	defer unmount()

	unpacked, err := unpackCombinedDtb(exec, b.BaseDtbFullPath, mountPoint)
	if err != nil {
		return err
	}
	defer func() { _ = exec.Remove(unpacked.monza) }()

	temporaryDtb := mountPoint.Join(temporaryDtbName())
	defer func() { _ = exec.Remove(temporaryDtb) }()

	args := buildOverlayCommand(b.OverlaysDir, unpacked.monza.String(), temporaryDtb, uniqueOverlays(overlays))
	if err := exec.Run(ctx, args...); err != nil {
		return err
	}

	packedDtb, err := packCombinedDtb(exec, temporaryDtb, unpacked)
	if err != nil {
		return err
	}

	return moveDeviceTree(exec, packedDtb, mountPoint.Join(b.DtbFileName))
}

func (b VentunoQAddon) Apply(ctx context.Context, exec executor.Executor, overlays []string) error {
	workDir := paths.New("/run/arduino-linux-config")
	mountPoint := workDir.Join("dtb")
	if err := exec.MkdirAll(mountPoint); err != nil {
		return fmt.Errorf("failed to create mountPoint: %w", err)
	}

	unmount, err := mountDeviceTree(ctx, exec, "/dev/disk/by-partlabel/dtb_a", mountPoint.String())
	if err != nil {
		return err
	}
	defer unmount()

	var plugged []*paths.Path
	defer func() {
		for _, file := range plugged {
			_ = exec.Remove(file)
		}
	}()

	current := b.AddonDir.Join(b.BaseDtbFile)
	for i, overlay := range uniqueOverlays(overlays) {
		addon, connector, found := strings.Cut(overlay, "@")
		if !found {
			return fmt.Errorf("invalid addon %q: expected <addon>@<connector>", overlay)
		}

		next := workDir.Join(fmt.Sprintf("%s.%d", temporaryDtbName(), i))
		plugged = append(plugged, next)
		if err := exec.Run(ctx, b.AddonDir.Join("fdtaddon").String(),
			"-i", current.String(), "-t", connector, "-o", next.String(), b.AddonDir.Join(addon).String()); err != nil {
			return err
		}
		current = next
	}

	// fdtaddon writes a v20 dtb, the bootloader reads only v17
	temporaryDtb := mountPoint.Join(temporaryDtbName())
	defer func() { _ = exec.Remove(temporaryDtb) }()
	if err := exec.Run(ctx, b.AddonDir.Join("dtc").String(),
		"-q", "-I", "dtb", "-O", "dtb", "-V", "17", "-o", temporaryDtb.String(), current.String()); err != nil {
		return err
	}

	return moveDeviceTree(exec, temporaryDtb, mountPoint.Join(b.DtbFileName))
}

func uniqueOverlays(overlays []string) []string {
	slices.Sort(overlays)
	return slices.Compact(overlays)
}

// The nanosecond timestamp keeps concurrent instances apart.
func temporaryDtbName() string {
	return fmt.Sprintf("temporaryDeviceTree.%d.temp", time.Now().UnixNano())
}
