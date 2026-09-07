// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

// Package devicetree rebuilds the device tree of the whole board. The board has
// one device tree, so every enabled mount must be collected at every change.
package devicetree

import (
	"context"
	"fmt"
	"slices"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/overlay"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

// Desired holds the state requested by the user. A mount that is not in the map
// keeps the state stored on disk.
type Desired map[registry.MountName]status.MountStatus

// Outcome describes the result of a Rebuild: overlays dropped because
// incompatible with the selection, and whether a reboot is required for the
// new configuration to take effect.
type Outcome struct {
	Incompatible   []string
	RebootRequired bool
}

// Rebuild regenerates the device tree from every mount of the board and then
// stores the requested changes. With an empty Desired it reloads the state on
// disk without any change. The Outcome is fully populated even when the error
// is non-nil, so a dry-run caller can degrade gracefully.
func Rebuild(ctx context.Context, exec executor.Executor, reg registry.Registry, cfg config.Configuration, desired Desired) (Outcome, error) {
	overlays := make([]string, 0, len(reg.Mounts))
	var currentOverlays []string
	var incompatible []string
	for _, mount := range reg.Mounts {
		current, next, err := status.Get(cfg, mount)
		if err != nil {
			return Outcome{}, fmt.Errorf("failed to get status for %s: %w", mount.Name, err)
		}
		curFiles, _ := overlay.CollectForStatus(mount, current)
		currentOverlays = append(currentOverlays, curFiles...)

		state := next
		if d, requested := desired[mount.Name]; requested {
			state = d
		}
		files, removed := overlay.CollectForStatus(mount, state)
		overlays = append(overlays, files...)
		incompatible = append(incompatible, removed...)
	}

	outcome := Outcome{
		Incompatible:   incompatible,
		RebootRequired: !slices.Equal(sortedUnique(currentOverlays), sortedUnique(overlays)),
	}

	applier, err := config.GetBoard()
	if err != nil {
		return outcome, err
	}

	if err := applier.Apply(ctx, exec, overlays); err != nil {
		return outcome, err
	}

	// The registry order keeps the written files, and so the reported effects,
	// the same on every run.
	for _, mount := range reg.Mounts {
		state, requested := desired[mount.Name]
		if !requested {
			continue
		}
		if err := status.Update(exec, cfg, mount, state); err != nil {
			return outcome, fmt.Errorf("failed to update status for %s: %w", mount.Name, err)
		}
	}
	return outcome, nil
}

func sortedUnique(files []string) []string {
	out := slices.Clone(files)
	slices.Sort(out)
	return slices.Compact(out)
}
