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

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/overlay"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

// Desired holds the state requested by the user. A mount that is not in the map
// keeps the state stored on disk.
type Desired map[registry.MountName]status.MountStatus

// Rebuild regenerates the device tree from every mount of the board and then
// stores the requested changes. With an empty Desired it reloads the state on
// disk without any change. It also returns the base overlays dropped because
// they were incompatible with the selection.
func Rebuild(ctx context.Context, exec executor.Executor, reg registry.Registry, cfg config.Configuration, desired Desired) ([]string, error) {
	applier, err := config.GetBoard()
	if err != nil {
		return nil, err
	}

	overlays := make([]string, 0, len(reg.Mounts))
	var incompatible []string
	for _, mount := range reg.Mounts {
		state, requested := desired[mount.Name]
		if !requested {
			if _, state, err = status.Get(cfg, mount); err != nil {
				return nil, fmt.Errorf("failed to get status for %s: %w", mount.Name, err)
			}
		}
		files, removed := overlay.CollectForStatus(mount, state)
		overlays = append(overlays, files...)
		incompatible = append(incompatible, removed...)
	}

	if err := applier.Apply(ctx, exec, overlays); err != nil {
		return incompatible, err
	}

	// The registry order keeps the written files, and so the reported effects,
	// the same on every run.
	for _, mount := range reg.Mounts {
		state, requested := desired[mount.Name]
		if !requested {
			continue
		}
		if err := status.Update(exec, cfg, mount, state); err != nil {
			return incompatible, fmt.Errorf("failed to update status for %s: %w", mount.Name, err)
		}
	}
	return incompatible, nil
}
