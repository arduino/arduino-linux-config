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

// CompleteMountName provides completion for the names of every part.
func CompleteMountName(reg registry.Registry, partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	completions := make([]cobra.Completion, 0, len(reg.Mounts))
	for _, mount := range reg.Mounts {
		completions = append(completions, cobra.Completion(mount.Name))
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// CompleteDeviceOption provides completion for device options in the format "device=option".
// It excludes devices that have already been specified in previous args.
func CompleteDeviceOption(mount registry.Mount, args []string, partial string) ([]cobra.Completion, cobra.ShellCompDirective) {
	// Build a set of device names already used in previous args.
	usedDevices := make(map[registry.DeviceName]struct{}, len(args))
	for _, arg := range args {
		if name, _, found := strings.Cut(arg, "="); found {
			usedDevices[registry.DeviceName(name)] = struct{}{}
		}
	}

	if !strings.Contains(partial, "=") {
		// Suggest device names for the first part of the argument, excluding already-used ones.
		completions := make([]cobra.Completion, 0, len(mount.Devices))
		for _, device := range mount.Devices {
			if _, used := usedDevices[device.Name]; !used {
				completions = append(completions, cobra.Completion(device.Name+"="))
			}
		}
		return completions, cobra.ShellCompDirectiveNoSpace
	}

	// Suggest options for the specified device, excluding already-used devices.
	completions := make([]cobra.Completion, 0, len(mount.Devices)*4)
	for _, device := range mount.Devices {
		if _, used := usedDevices[device.Name]; used {
			continue
		}
		for _, option := range device.Options {
			completions = append(completions, fmt.Sprintf("%s=%s", device.Name, option.Name))
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
