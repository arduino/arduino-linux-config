// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package completion

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/internal/registry"
)

// CompleteCarrierName provides completion for carrier names on the first argument.
func CompleteCarrierName(registry registry.Registry, args []string, partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	completions := make([]cobra.Completion, 0, len(registry.Carriers))
	for _, carrier := range registry.Carriers {
		completions = append(completions, cobra.Completion(carrier.Name))
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// CompleteDeviceOption provides completion for device options in the format "device=option".
// It excludes devices that have already been specified in previous args.
func CompleteDeviceOption(carrier registry.Carrier, args []string, partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	// Build a set of device names already used in previous args.
	usedDevices := make(map[registry.CarrierDeviceName]struct{}, len(args))
	for _, arg := range args {
		if name, _, found := strings.Cut(arg, "="); found {
			usedDevices[registry.CarrierDeviceName(name)] = struct{}{}
		}
	}

	if !strings.Contains(partial, "=") {
		// Suggest device names for the first part of the argument, excluding already-used ones.
		completions := make([]cobra.Completion, 0, len(carrier.Devices))
		for _, device := range carrier.Devices {
			if _, used := usedDevices[device.Name]; !used {
				completions = append(completions, cobra.Completion(device.Name+"="))
			}
		}
		return completions, cobra.ShellCompDirectiveNoSpace
	}

	// Suggest options for the specified device, excluding already-used devices.
	completions := make([]cobra.Completion, 0, len(carrier.Devices)*4)
	for _, device := range carrier.Devices {
		if _, used := usedDevices[device.Name]; used {
			continue
		}
		for _, option := range device.Options {
			completions = append(completions, fmt.Sprintf("%s=%s", device.Name, option.Name))
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
