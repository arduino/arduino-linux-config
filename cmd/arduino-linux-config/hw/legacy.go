// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"github.com/spf13/cobra"

	"github.com/arduino/arduino-linux-config/cmd/feedback"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

// The JSON of the "carrier" group keeps the shape of the v0.2.x releases, so
// the tools that read it keep working. Only the carriers are reported, because
// the old releases knew nothing about the hats.

func newLegacyListCmd(reg registry.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the carriers available for this board",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			feedback.PrintResult(buildLegacyListResult(reg))
		},
	}
}

func buildLegacyListResult(reg registry.Registry) legacyListResult {
	return legacyListResult{inner: buildListResult(reg.ByKind(registry.KindCarrier))}
}

type legacyListResult struct {
	inner listResult
}

func (r legacyListResult) String() string {
	return r.inner.String()
}

func (r legacyListResult) Data() interface{} {
	return legacyListData(r.inner.Mounts)
}

type legacyCarriersResult struct {
	Carriers []legacyCarrierResult `json:"carriers"`
}

type legacyCarrierResult struct {
	Name    string                      `json:"name"`
	Devices []legacyCarrierDeviceResult `json:"devices"`
}

type legacyCarrierDeviceResult struct {
	Name             string   `json:"name"`
	DeviceType       string   `json:"device_type"`
	AvailableDevices []string `json:"available_devices"`
}

func legacyListData(mounts []listMount) legacyCarriersResult {
	result := legacyCarriersResult{Carriers: make([]legacyCarrierResult, 0, len(mounts))}
	for _, mount := range mounts {
		devices := make([]legacyCarrierDeviceResult, 0, len(mount.Devices))
		for _, device := range mount.Devices {
			devices = append(devices, legacyCarrierDeviceResult{
				Name:             device.Name,
				DeviceType:       device.DeviceType,
				AvailableDevices: device.Options,
			})
		}
		result.Carriers = append(result.Carriers, legacyCarrierResult{
			Name:    mount.Name,
			Devices: devices,
		})
	}
	return result
}

type legacyShowResult struct {
	Carriers []legacyShowCarrierResult `json:"carriers"`
}

type legacyShowCarrierResult struct {
	CarrierName    string         `json:"carrier_name"`
	CurrentEnabled bool           `json:"current_enabled"`
	NextEnabled    bool           `json:"next_enabled"`
	CurrentDevices []deviceResult `json:"current"`
	NextDevices    []deviceResult `json:"next"`
}

func legacyShowData(mounts []showMount) legacyShowResult {
	result := legacyShowResult{Carriers: make([]legacyShowCarrierResult, 0, len(mounts))}
	for _, mount := range mounts {
		result.Carriers = append(result.Carriers, legacyShowMount(mount))
	}
	return result
}

func legacyShowMount(mount showMount) legacyShowCarrierResult {
	return legacyShowCarrierResult{
		CarrierName:    mount.Name,
		CurrentEnabled: mount.CurrentEnabled,
		NextEnabled:    mount.NextEnabled,
		CurrentDevices: mount.CurrentDevices,
		NextDevices:    mount.NextDevices,
	}
}
