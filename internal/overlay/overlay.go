// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

// Package overlay centralizes the logic that translates a carrier configuration
// into the list of device tree overlay (dtbo) files that must be applied
package overlay

import (
	"slices"

	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

// Collect returns the overlay files required to enable a carrier according to the user selection and the base overlays that should be removed because they are incompatible with the selected options.
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
