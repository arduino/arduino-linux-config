// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

var showTestResult = showResult{Mounts: []showMount{{
	Name:           "media-carrier",
	Kind:           "carrier",
	CurrentEnabled: false,
	NextEnabled:    true,
	CurrentDevices: []deviceResult{},
	NextDevices:    []deviceResult{{Device: "display", Option: "5-dsi-touch-a", DeviceType: "display"}},
}}}

func TestShowDataKeepsTheNewShapeForHw(t *testing.T) {
	data, err := json.Marshal(showTestResult.Data())
	require.NoError(t, err)
	require.JSONEq(t, `{
		"mounts": [
			{
				"name": "media-carrier",
				"kind": "carrier",
				"current_enabled": false,
				"next_enabled": true,
				"current": [],
				"next": [{"device": "display", "option": "5-dsi-touch-a", "device_type": "display"}]
			}
		]
	}`, string(data))
}

// The v0.2.x JSON groups the mounts under "carriers" and names them
// "carrier_name", with no kind.
func TestShowDataKeepsTheOldShapeForCarrier(t *testing.T) {
	legacyResult := legacyShow{inner: showTestResult}

	data, err := json.Marshal(legacyResult.Data())
	require.NoError(t, err)
	require.JSONEq(t, `{
		"carriers": [
			{
				"carrier_name": "media-carrier",
				"current_enabled": false,
				"next_enabled": true,
				"current": [],
				"next": [{"device": "display", "option": "5-dsi-touch-a", "device_type": "display"}]
			}
		]
	}`, string(data))
}

// The v0.2.x enable and disable reported the affected carrier out of any list.
func TestShowDataOfASingleCarrierIsNotWrapped(t *testing.T) {
	legacyResult := legacyShow{inner: showTestResult, single: true}

	data, err := json.Marshal(legacyResult.Data())
	require.NoError(t, err)
	require.JSONEq(t, `{
		"carrier_name": "media-carrier",
		"current_enabled": false,
		"next_enabled": true,
		"current": [],
		"next": [{"device": "display", "option": "5-dsi-touch-a", "device_type": "display"}]
	}`, string(data))
}
