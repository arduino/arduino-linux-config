// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/hw/completion"
	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
)

func newShowCmd(reg registry.Registry, cfg config.Configuration) *cobra.Command {
	return &cobra.Command{
		Use:   "show [name]",
		Short: "Show the configuration of the board, or of one part of it",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mounts := reg.Mounts
			if len(args) > 0 {
				mounts = []registry.Mount{findMount(reg, args[0])}
			}
			showHandler(reg, cfg, mounts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completion.CompleteMountName(reg, toComplete)
		},
	}
}

func showHandler(reg registry.Registry, cfg config.Configuration, mounts []registry.Mount) {
	result := showResult{Mounts: make([]showMount, 0, len(mounts))}
	for _, mount := range mounts {
		current, next, err := status.Get(cfg, mount)
		if err != nil {
			feedback.Fatal(fmt.Sprintf("failed to get status for %s: %v", mount.Name, err), feedback.ErrGeneric)
		}
		result.Mounts = append(result.Mounts, showMount{
			Name:           string(mount.Name),
			Kind:           mount.Kind.Label(),
			CurrentEnabled: current.Enable,
			NextEnabled:    next.Enable,
			CurrentDevices: withDeviceType(mount, current.StatusDevices),
			NextDevices:    withDeviceType(mount, next.StatusDevices),
		})
	}
	feedback.PrintResult(result)
}

// findMount resolves a name over every kind, because the name alone tells the
// tool where the part plugs in.
func findMount(reg registry.Registry, name string) registry.Mount {
	mount, exist := reg.FindByName(name)
	if !exist {
		feedback.Fatal(fmt.Sprintf("%q not supported by this board", name), feedback.ErrBadArgument)
	}
	return mount
}

type showResult struct {
	Mounts []showMount `json:"mounts"`
}

type showMount struct {
	Name           string         `json:"name"`
	Kind           string         `json:"kind"`
	CurrentEnabled bool           `json:"current_enabled"`
	NextEnabled    bool           `json:"next_enabled"`
	CurrentDevices []deviceResult `json:"current"`
	NextDevices    []deviceResult `json:"next"`
}

type deviceResult struct {
	Device     string `json:"device"`
	Option     string `json:"option"`
	DeviceType string `json:"device_type"`
}

func withDeviceType(mount registry.Mount, devices []status.StatusDevice) []deviceResult {
	result := make([]deviceResult, 0, len(devices))
	for _, device := range devices {
		registered, _ := mount.FindDeviceByName(registry.DeviceName(device.Device))
		result = append(result, deviceResult{
			Device:     device.Device,
			Option:     device.Option,
			DeviceType: string(registered.DeviceType),
		})
	}
	return result
}

func (r showResult) String() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	for _, mount := range r.Mounts {
		fmt.Fprintf(w, "%s\t%s\t[current: %s]\t[next: %s]\n",
			mount.Kind, mount.Name, enabledLabel(mount.CurrentEnabled), enabledLabel(mount.NextEnabled))

		for i, device := range mount.CurrentDevices {
			next := "none"
			if i < len(mount.NextDevices) {
				next = mount.NextDevices[i].Option
			}
			fmt.Fprintf(w, "\t  %s:\t[current: %s]\t[next: %s]\n", device.Device, device.Option, next)
		}
	}

	w.Flush()
	return b.String()
}

func (r showResult) Data() any {
	return r
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
