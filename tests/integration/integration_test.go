// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build integration

package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type deviceResult struct {
	Device     string `json:"device"`
	Option     string `json:"option"`
	DeviceType string `json:"device_type"`
}

type carrierResult struct {
	CarrierName    string         `json:"carrier_name"`
	CurrentEnabled bool           `json:"current_enabled"`
	NextEnabled    bool           `json:"next_enabled"`
	Current        []deviceResult `json:"current"`
	Next           []deviceResult `json:"next"`
	Warnings       []string       `json:"warnings"`
}

type showResult struct {
	Carriers []carrierResult `json:"carriers"`
}

func TestCarrierShowEmpty(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out := execInContainer(t, "arduino-linux-config", "carrier", "show", "--format", "json")

	var result showResult
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON")

	require.Len(t, result.Carriers, 1)

	carrier := result.Carriers[0]
	require.Equal(t, "media-carrier", carrier.CarrierName)
	require.Equal(t, false, carrier.CurrentEnabled)
	require.Equal(t, false, carrier.NextEnabled)

	expectedDevices := []deviceResult{
		{Device: "camera0", Option: "none", DeviceType: "camera"},
		{Device: "camera1", Option: "none", DeviceType: "camera"},
		{Device: "display", Option: "none", DeviceType: "display"},
	}

	require.Equal(t, expectedDevices, carrier.Current)
	require.Equal(t, expectedDevices, carrier.Next)
}

func TestCarrierEnableAllDevices(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out := execInContainer(t, "arduino-linux-config", "carrier", "enable", "media-carrier",
		"camera0=type1-2lanes",
		"camera1=type1-2lanes",
		"display=8-dsi-touch-a",
		"--format", "json",
	)

	var result carrierResult
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON")

	require.Equal(t, "media-carrier", result.CarrierName)
	require.Equal(t, false, result.CurrentEnabled)
	require.Equal(t, true, result.NextEnabled)

	expectedCurrent := []deviceResult{
		{Device: "camera0", Option: "none", DeviceType: "camera"},
		{Device: "camera1", Option: "none", DeviceType: "camera"},
		{Device: "display", Option: "none", DeviceType: "display"},
	}

	expectedNext := []deviceResult{
		{Device: "camera0", Option: "type1-2lanes", DeviceType: "camera"},
		{Device: "camera1", Option: "type1-2lanes", DeviceType: "camera"},
		{Device: "display", Option: "8-dsi-touch-a", DeviceType: "display"},
	}

	require.Equal(t, expectedCurrent, result.Current)
	require.Equal(t, expectedNext, result.Next)
}

func TestCarrierDisable(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	// First enable all devices
	execInContainer(t, "arduino-linux-config", "carrier", "enable", "media-carrier",
		"camera0=type1-2lanes",
		"camera1=type1-2lanes",
		"display=8-dsi-touch-a",
	)

	// Then disable
	out := execInContainer(t, "arduino-linux-config", "carrier", "disable", "media-carrier", "--format", "json")

	var result carrierResult
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON")

	require.Equal(t, "media-carrier", result.CarrierName)
	require.Equal(t, false, result.CurrentEnabled)
	require.Equal(t, false, result.NextEnabled)

	expectedDevices := []deviceResult{
		{Device: "camera0", Option: "none", DeviceType: "camera"},
		{Device: "camera1", Option: "none", DeviceType: "camera"},
		{Device: "display", Option: "none", DeviceType: "display"},
	}

	require.Equal(t, expectedDevices, result.Current)
	require.Equal(t, expectedDevices, result.Next)
}

type listDeviceResult struct {
	Name             string   `json:"name"`
	DeviceType       string   `json:"device_type"`
	AvailableDevices []string `json:"available_devices"`
}

type listCarrierResult struct {
	Name    string             `json:"name"`
	Devices []listDeviceResult `json:"devices"`
}

type listResult struct {
	Carriers []listCarrierResult `json:"carriers"`
}

func TestCarrierList(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out := execInContainer(t, "arduino-linux-config", "carrier", "list", "--format", "json")

	var result listResult
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON")

	require.Len(t, result.Carriers, 1)

	carrier := result.Carriers[0]
	require.Equal(t, "media-carrier", carrier.Name)
	require.Len(t, carrier.Devices, 3)

	require.Equal(t, listDeviceResult{
		Name:             "camera0",
		DeviceType:       "camera",
		AvailableDevices: []string{"none", "type1-2lanes", "type1-4lanes"},
	}, carrier.Devices[0])

	require.Equal(t, listDeviceResult{
		Name:             "camera1",
		DeviceType:       "camera",
		AvailableDevices: []string{"none", "type1-2lanes", "type1-4lanes"},
	}, carrier.Devices[1])

	require.Equal(t, listDeviceResult{
		Name:             "display",
		DeviceType:       "display",
		AvailableDevices: []string{"none", "5-dsi-touch-a", "8-dsi-touch-a", "10-dsi-touch-a"},
	}, carrier.Devices[2])
}

func TestWriteReadFromConfigFile(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	// Assert post installation config
	require.NotEmpty(t, execInContainer(t, "ls", "-a", "/var/lib/arduino-linux-config/status"))

	out := execInContainer(t, "arduino-linux-config", "carrier", "enable", "media-carrier",
		"camera1=type1-2lanes",
		"--format", "json",
	)

	var result carrierResult
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON")

	// Assert configuration file created after the first configuration
	configFile := "/var/lib/arduino-linux-config/status/media-carrier.json"
	require.NotEmpty(t, execInContainer(t, "ls", configFile))

	require.Equal(t, "media-carrier", result.CarrierName)
	require.Equal(t, false, result.CurrentEnabled)
	require.Equal(t, true, result.NextEnabled)

	expectedCurrent := []deviceResult{
		{Device: "camera0", Option: "none", DeviceType: "camera"},
		{Device: "camera1", Option: "none", DeviceType: "camera"},
		{Device: "display", Option: "none", DeviceType: "display"},
	}

	expectedNext := []deviceResult{
		{Device: "camera0", Option: "none", DeviceType: "camera"},
		{Device: "camera1", Option: "type1-2lanes", DeviceType: "camera"},
		{Device: "display", Option: "none", DeviceType: "display"},
	}

	require.Equal(t, expectedCurrent, result.Current)
	require.Equal(t, expectedNext, result.Next)

	// Assert read from file
	out = execInContainer(t, "arduino-linux-config", "carrier", "enable", "media-carrier",
		"camera1=type1-2lanes",
		"--format", "json",
	)

	require.Equal(t, "media-carrier", result.CarrierName)
	require.Equal(t, false, result.CurrentEnabled)
	require.Equal(t, true, result.NextEnabled)

	require.Equal(t, expectedCurrent, result.Current)
	require.Equal(t, expectedNext, result.Next)
}
