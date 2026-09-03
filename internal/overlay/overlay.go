// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

// Package overlay centralizes the logic that translates a carrier configuration
// into the list of device tree overlay (dtbo) files that must be applied
package overlay

import (
	"fmt"
	"slices"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

// Collect returns the overlay files required to enable a carrier according to the user selection.
// It also returns the base overlays that were removed because they were incompatible with the
// selected options.
func Collect(carrier registry.Carrier, userSelection []status.StatusDevice) ([]string, []string) {
	var baseFiles, dtboFiles, incompatibleFiles []string

	for _, selection := range userSelection {
		device, exist := carrier.FindDeviceByName(registry.CarrierDeviceName(selection.Device))
		if !exist {
			continue
		}
		for _, option := range device.Options {
			// get the user selected option and collect incompatibilities
			if option.Name == selection.Option {
				dtboFiles = append(dtboFiles, option.DtboFiles...)
				incompatibleFiles = append(incompatibleFiles, option.IncompatibleDtbo...)
				break
			}
		}
	}

	// collect base dtbo files for the media carrier.
	baseFiles = append(baseFiles, carrier.EnabledDtbos...)

	// check for incompatible overlays in the basic configuration
	// in this case, the basic overlay can be removed in favor of the device overlays
	incompatibleOverlays := getIntersection(baseFiles, incompatibleFiles)

	// remove incompatible layer and proceed
	if len(incompatibleOverlays) > 0 {
		baseFiles = slices.DeleteFunc(baseFiles, func(overlay string) bool {
			return slices.Contains(incompatibleFiles, overlay)
		})
		dtboFiles = slices.DeleteFunc(dtboFiles, func(overlay string) bool {
			return slices.Contains(incompatibleFiles, overlay)
		})
	}

	return append(dtboFiles, baseFiles...), incompatibleOverlays
}

// Returns the overlay files needed to restore a carrier to its disabled state.
func CollectDisabled(carrier registry.Carrier) []string {
	baseFiles := make([]string, 0)

	for _, device := range carrier.Devices {
		for _, option := range device.Options {
			if option.Name == string(registry.None) {
				baseFiles = append(baseFiles, option.DtboFiles...)
			}
		}
	}

	// Add disable dtbs to restore original board configuration.
	baseFiles = append(baseFiles, carrier.DisabledDtbos...)
	return baseFiles
}

// Resolves the overlays for a carrier from its persisted status.
func CollectForStatus(carrier registry.Carrier, current status.CarrierStatus) []string {
	if !current.Enable {
		return CollectDisabled(carrier)
	}
	files, _ := Collect(carrier, current.StatusDevices)
	return files
}

func getIntersection(a, b []string) []string {
	var result []string
	for _, v := range a {
		if slices.Contains(b, v) {
			result = append(result, v)
		}
	}
	slices.Sort(result)
	return slices.Compact(result)
}

func GetDtboForAddon(addons []registry.Addon, addonName registry.AddonName) []string {
	for _, addon := range addons {
		if addon.Name == addonName {
			return addon.EnabledDtbos
		}
	}
	return []string{}
}

// GetConfiguredCarriersOverlay returns the overlays required by the persisted
// carrier configuration for the next boot and the reloaded carriers.
func GetConfiguredCarriersOverlay(cfg config.Configuration, reg registry.Registry) ([]string, []string) {
	overlays := make([]string, 0, len(reg.Carriers))
	carriers := make([]string, 0, len(reg.Carriers))
	for _, carrier := range reg.Carriers {
		// A carrier without a persisted status is reported as disabled
		_, next, err := status.Get(cfg, carrier)
		if err != nil {
			feedback.Fatal(fmt.Sprintf("failed to get status for carrier %s: %v", carrier.Name, err), feedback.ErrGeneric)
		}
		overlays = append(overlays, CollectForStatus(carrier, next)...)
		carriers = append(carriers, string(carrier.Name))
	}
	return overlays, carriers
}

// GetConfiguredAddonsOverlay returns the overlays required by the persisted
// addon configuration for the current boot and its name
func GetConfiguredAddonsOverlay(cfg config.Configuration, reg registry.Registry) ([]string, string) {
	nextAddonName, err := status.GetNextConfiguredAddon(cfg, reg)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("failed to get addons status: %v", err), feedback.ErrGeneric)
	}
	return GetDtboForAddon(reg.Addons, nextAddonName), string(nextAddonName)
}
