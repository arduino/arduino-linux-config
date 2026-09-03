// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package status

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

// addonsStatusFileName is the single file that stores the state of every addon.
const addonsStatusFileName = "addons"

// AddonsStatusFile mirrors the carrier status file layout but stores every
// addon in a single file. Since an addon has no configurable devices, each
// entry only tracks the enabled flag for the current and the next boot.
type AddonsStatusFile struct {
	CurrentStatus AddonsStatus `json:"current_status"`
	NextStatus    AddonsStatus `json:"next_status"`
}

type AddonsStatus struct {
	CreatedAt string                                 `json:"created_at"`
	Addons    map[registry.AddonName]AddonStatusInfo `json:"addons"`
}

type AddonStatusInfo struct {
	Enabled   bool   `json:"status"`
	CreatedAt string `json:"created_at"`
}

// AddonState is the resolved current/next enabled state for a single addon.
type AddonState struct {
	Name    registry.AddonName
	Current bool
	Next    bool
}

// GetAllAddons returns the resolved current/next state for every addon in the
// registry. It does not persist any change on disk because show runs as a
// non-root user.
func GetAllAddons(cfg config.Configuration, reg registry.Registry) ([]AddonState, error) {
	status, err := loadAddonsStatusFile(getAddonsStatusFile(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to load status file %v", err)
	}

	currentBootId, err := getCurrentBootID()
	if err != nil {
		return nil, fmt.Errorf("failed to get boot-id: %v", err)
	}

	return resolveAddons(status, reg, currentBootId), nil
}

// GetNextConfiguredAddon returns the addon configured for the next boot, or
// an empty name if every addon is disabled.
func GetNextConfiguredAddon(cfg config.Configuration, reg registry.Registry) (registry.AddonName, error) {
	states, err := GetAllAddons(cfg, reg)
	if err != nil {
		return "", err
	}

	for _, state := range states {
		if state.Next {
			return state.Name, nil
		}
	}

	return "", nil
}

// UpdateAddons persists the enable/disable request described by changes. As for
// carriers, the request is stored in next_status and only becomes current_status
// after a reboot (tracked via the boot-id)
func UpdateAddons(exec executor.Executor, cfg config.Configuration, reg registry.Registry, changes map[registry.AddonName]bool) error {
	status, err := loadAddonsStatusFile(getAddonsStatusFile(cfg))
	if err != nil {
		return fmt.Errorf("failed to load status file %w", err)
	}

	currentBootId, err := getCurrentBootID()
	if err != nil {
		return fmt.Errorf("failed to get boot-id: %w", err)
	}

	applyAddonChanges(status, reg, changes, currentBootId)

	if err := saveAddonsStatusFile(exec, getAddonsStatusFile(cfg), *status); err != nil {
		return fmt.Errorf("failed to save status file: %w", err)
	}
	return nil
}

// applyAddonChanges promotes the outdated next_status into current_status and
// records the requested enable/disable changes in next_status, guaranteeing that
// every registry addon is present in both sections.
func applyAddonChanges(status *AddonsStatusFile, reg registry.Registry, changes map[registry.AddonName]bool, currentBootId string) {
	// Promote any next_status written in a previous boot into current_status.
	resolveAddons(status, reg, currentBootId)

	for _, addon := range reg.Addons {
		current := status.CurrentStatus.Addons[addon.Name]
		current.CreatedAt = currentBootId
		status.CurrentStatus.Addons[addon.Name] = current

		next := status.NextStatus.Addons[addon.Name]
		if enable, ok := changes[addon.Name]; ok {
			next.Enabled = enable
		} else {
			// Only one addon at time is allowed
			next.Enabled = false
		}
		next.CreatedAt = currentBootId
		status.NextStatus.Addons[addon.Name] = next
	}
	status.CurrentStatus.CreatedAt = currentBootId
	status.NextStatus.CreatedAt = currentBootId
}

// resolveAddons applies the boot-id fixup: if an addon next_status was written
// in a previous boot it has already taken effect, so it is promoted to
// current_status. It returns the resolved state for every registry addon in
// registry order.
func resolveAddons(status *AddonsStatusFile, reg registry.Registry, currentBootId string) []AddonState {
	states := make([]AddonState, 0, len(reg.Addons))
	for _, addon := range reg.Addons {
		next := status.NextStatus.Addons[addon.Name]
		if next.CreatedAt != currentBootId {
			status.CurrentStatus.Addons[addon.Name] = next
		}
		current := status.CurrentStatus.Addons[addon.Name]
		states = append(states, AddonState{
			Name:    addon.Name,
			Current: current.Enabled,
			Next:    next.Enabled,
		})
	}
	return states
}

func getAddonsStatusFile(cfg config.Configuration) *paths.Path {
	return cfg.StatusDir().Join(addonsStatusFileName + ".json")
}

func loadAddonsStatusFile(statusFile *paths.Path) (*AddonsStatusFile, error) {
	data, err := statusFile.ReadFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newAddonsStatusFile(), nil
		}
		return nil, err
	}

	var status AddonsStatusFile
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("could not parse json: %w", err)
	}
	if status.CurrentStatus.Addons == nil {
		status.CurrentStatus.Addons = make(map[registry.AddonName]AddonStatusInfo)
	}
	if status.NextStatus.Addons == nil {
		status.NextStatus.Addons = make(map[registry.AddonName]AddonStatusInfo)
	}
	return &status, nil
}

func newAddonsStatusFile() *AddonsStatusFile {
	return &AddonsStatusFile{
		CurrentStatus: AddonsStatus{Addons: make(map[registry.AddonName]AddonStatusInfo)},
		NextStatus:    AddonsStatus{Addons: make(map[registry.AddonName]AddonStatusInfo)},
	}
}

func saveAddonsStatusFile(exec executor.Executor, statusFile *paths.Path, status AddonsStatusFile) error {
	data, err := json.MarshalIndent(status, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	// nolint:gosec // G306: Status file must be readable
	if err := exec.WriteFile(statusFile, data, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}
