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

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

func newListCmd(reg registry.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the carriers and the addons available for this board",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			feedback.PrintResult(buildListResult(reg))
		},
	}
}

func buildListResult(reg registry.Registry) listResult {
	result := listResult{Mounts: make([]listMount, 0, len(reg.Mounts))}
	for _, mount := range reg.Mounts {
		devices := make([]listDevice, 0, len(mount.Devices))
		for _, device := range mount.Devices {
			options := make([]string, len(device.Options))
			for i, option := range device.Options {
				options[i] = option.Name
			}
			devices = append(devices, listDevice{
				Name:       string(device.Name),
				DeviceType: string(device.DeviceType),
				Options:    options,
			})
		}
		result.Mounts = append(result.Mounts, listMount{
			Name:    string(mount.Name),
			Kind:    string(mount.Kind),
			Devices: devices,
		})
	}
	return result
}

type listResult struct {
	Mounts []listMount `json:"mounts"`
}

type listMount struct {
	Name    string       `json:"name"`
	Kind    string       `json:"kind"`
	Devices []listDevice `json:"devices"`
}

type listDevice struct {
	Name       string   `json:"name"`
	DeviceType string   `json:"device_type"`
	Options    []string `json:"options"`
}

// The parts are grouped by connector, because a carrier and an addon plug into
// different places on the board.
func (r listResult) String() string {
	var b strings.Builder

	for _, group := range []struct {
		kind    registry.Kind
		title   string
		comment string
	}{
		{registry.KindCarrier, "CARRIERS", "connect to the carrier connector"},
		{registry.KindHat, "ADDONS", "connect to the 40-pin hat connector"},
	} {
		mounts := make([]listMount, 0, len(r.Mounts))
		for _, mount := range r.Mounts {
			if mount.Kind == string(group.kind) {
				mounts = append(mounts, mount)
			}
		}
		if len(mounts) == 0 {
			continue
		}

		fmt.Fprintf(&b, "%s (%s, one at a time)\n", group.title, group.comment)
		w := tabwriter.NewWriter(&b, 0, 0, 4, ' ', 0)
		for _, mount := range mounts {
			if len(mount.Devices) == 0 {
				fmt.Fprintf(w, "  %s\n", mount.Name)
				continue
			}
			for i, device := range mount.Devices {
				name := ""
				if i == 0 {
					name = mount.Name
				}
				fmt.Fprintf(w, "  %s\t%s\t%s\n", name, device.Name, strings.Join(device.Options, ", "))
			}
		}
		w.Flush()
		fmt.Fprintln(&b)
	}

	return b.String()
}

func (r listResult) Data() interface{} {
	return r
}
