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

type mountResult struct {
	Name           string         `json:"name"`
	Kind           string         `json:"kind"`
	CurrentEnabled bool           `json:"current_enabled"`
	NextEnabled    bool           `json:"next_enabled"`
	Current        []deviceResult `json:"current"`
	Next           []deviceResult `json:"next"`
}

type showResult struct {
	Mounts []mountResult `json:"mounts"`
}

// Every command prints the whole board, so the tests select the mount to assert.
func mountByName(t *testing.T, out, name string) mountResult {
	var result showResult
	require.NoError(t, json.Unmarshal([]byte(out), &result), "output should be valid JSON")
	for _, mount := range result.Mounts {
		if mount.Name == name {
			return mount
		}
	}
	t.Fatalf("mount %q not found in %s", name, out)
	return mountResult{}
}

func TestMountShowEmpty(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out := execInContainer(t, "arduino-linux-config", "hw", "show", "--format", "json")

	carrier := mountByName(t, out, "media-carrier")
	require.Equal(t, "carrier", carrier.Kind)
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

func TestMountEnableAllDevices(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out := execInContainer(t, "arduino-linux-config", "hw", "enable", "media-carrier",
		"camera0=type1-2lanes",
		"camera1=type1-2lanes",
		"display=8-dsi-touch-a",
		"--format", "json",
	)

	result := mountByName(t, out, "media-carrier")
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

func TestMountDisable(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	// First enable all devices
	execInContainer(t, "arduino-linux-config", "hw", "enable", "media-carrier",
		"camera0=type1-2lanes",
		"camera1=type1-2lanes",
		"display=8-dsi-touch-a",
	)

	// Then disable
	out := execInContainer(t, "arduino-linux-config", "hw", "disable", "media-carrier", "--format", "json")

	result := mountByName(t, out, "media-carrier")
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
	Name       string   `json:"name"`
	DeviceType string   `json:"device_type"`
	Options    []string `json:"options"`
}

type listMountResult struct {
	Name    string             `json:"name"`
	Kind    string             `json:"kind"`
	Devices []listDeviceResult `json:"devices"`
}

type listResult struct {
	Mounts []listMountResult `json:"mounts"`
}

func TestMountList(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out := execInContainer(t, "arduino-linux-config", "hw", "list", "--format", "json")

	var result listResult
	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err, "output should be valid JSON")

	require.Len(t, result.Mounts, 1)

	carrier := result.Mounts[0]
	require.Equal(t, "carrier", carrier.Kind)
	require.Equal(t, "media-carrier", carrier.Name)
	require.Len(t, carrier.Devices, 3)

	require.Equal(t, listDeviceResult{
		Name:       "camera0",
		DeviceType: "camera",
		Options:    []string{"none", "type1-2lanes", "type1-4lanes"},
	}, carrier.Devices[0])

	require.Equal(t, listDeviceResult{
		Name:       "camera1",
		DeviceType: "camera",
		Options:    []string{"none", "type1-2lanes", "type1-4lanes"},
	}, carrier.Devices[1])

	require.Equal(t, listDeviceResult{
		Name:       "display",
		DeviceType: "display",
		Options:    []string{"none", "5-dsi-touch-a", "8-dsi-touch-a", "10-dsi-touch-a"},
	}, carrier.Devices[2])
}

func TestWriteReadFromConfigFile(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	// Assert post installation config
	require.NotEmpty(t, execInContainer(t, "ls", "-a", "/var/lib/arduino-linux-config/status"))

	out := execInContainer(t, "arduino-linux-config", "hw", "enable", "media-carrier",
		"camera1=type1-2lanes",
		"--format", "json",
	)

	result := mountByName(t, out, "media-carrier")

	// Assert configuration file created after the first configuration
	configFile := "/var/lib/arduino-linux-config/status/media-carrier.json"
	require.NotEmpty(t, execInContainer(t, "ls", configFile))

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
	out = execInContainer(t, "arduino-linux-config", "hw", "enable", "media-carrier",
		"camera1=type1-2lanes",
		"--format", "json",
	)
	result = mountByName(t, out, "media-carrier")

	require.Equal(t, false, result.CurrentEnabled)
	require.Equal(t, true, result.NextEnabled)

	require.Equal(t, expectedCurrent, result.Current)
	require.Equal(t, expectedNext, result.Next)
}
