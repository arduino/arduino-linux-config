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
		Short: "List the carriers and the hats available for this board",
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

// The parts are grouped by connector, because a carrier and a hat plug into
// different places on the board. Each group keeps the table of the command it
// replaces.
func (r listResult) String() string {
	var b strings.Builder

	for _, group := range []struct {
		kind   registry.Kind
		header string
	}{
		{registry.KindCarrier, "CARRIER"},
		{registry.KindHat, "HAT"},
	} {
		mounts := make([]listMount, 0, len(r.Mounts))
		withDevices := false
		for _, mount := range r.Mounts {
			if mount.Kind != string(group.kind) {
				continue
			}
			mounts = append(mounts, mount)
			withDevices = withDevices || len(mount.Devices) > 0
		}
		if len(mounts) == 0 {
			continue
		}
		if b.Len() > 0 {
			fmt.Fprintln(&b)
		}

		// minwidth: 0, tabwidth: 0, padding: 4, padchar: ' ', flags: 0
		w := tabwriter.NewWriter(&b, 0, 0, 4, ' ', 0)
		underline := strings.Repeat("-", len(group.header))
		if !withDevices {
			fmt.Fprintln(w, group.header)
			fmt.Fprintln(w, underline)
		} else {
			fmt.Fprintln(w, group.header+"\tDEVICE\tOPTIONS")
			fmt.Fprintln(w, underline+"\t------\t-------")
		}

		for i, mount := range mounts {
			if len(mount.Devices) == 0 {
				fmt.Fprintln(w, mount.Name)
				continue
			}
			// An empty row keeps the device blocks of two mounts apart.
			if i > 0 {
				fmt.Fprintln(w, "\t\t")
			}
			for j, device := range mount.Devices {
				name := ""
				if j == 0 {
					name = mount.Name
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", name, device.Name, strings.Join(device.Options, ", "))
			}
		}
		w.Flush()
	}

	return b.String()
}

func (r listResult) Data() interface{} {
	return r
}
