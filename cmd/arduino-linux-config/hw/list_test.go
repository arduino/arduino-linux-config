// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/registry"
)

var listTestRegistry = registry.Registry{Mounts: []registry.Mount{
	{
		Name: registry.MediaCarrier,
		Kind: registry.KindCarrier,
		Devices: []registry.Device{{
			Name:       registry.Display,
			DeviceType: registry.DeviceTypeDisplay,
			Options:    []registry.DeviceOption{{Name: "none"}, {Name: "5-dsi-touch-a"}},
		}},
	},
	{Name: registry.Automation, Kind: registry.KindHat},
}}

func TestListDataKeepsTheNewShapeForHw(t *testing.T) {
	data, err := json.Marshal(buildListResult(listTestRegistry, false).Data())
	require.NoError(t, err)
	require.JSONEq(t, `{
		"mounts": [
			{
				"name": "media-carrier",
				"kind": "carrier",
				"devices": [
					{"name": "display", "device_type": "display", "options": ["none", "5-dsi-touch-a"]}
				]
			},
			{"name": "automation", "kind": "hat", "devices": []}
		]
	}`, string(data))
}

// The v0.2.x JSON has no hat and no kind, and calls the options
// "available_devices".
func TestListDataKeepsTheOldShapeForCarrier(t *testing.T) {
	data, err := json.Marshal(buildListResult(listTestRegistry, true).Data())
	require.NoError(t, err)
	require.JSONEq(t, `{
		"carriers": [
			{
				"name": "media-carrier",
				"devices": [
					{"name": "display", "device_type": "display", "available_devices": ["none", "5-dsi-touch-a"]}
				]
			}
		]
	}`, string(data))
}
