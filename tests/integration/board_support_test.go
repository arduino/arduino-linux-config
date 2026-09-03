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

type addonsListResult struct {
	Addons []string `json:"addons"`
}

func carrierNames(t *testing.T, out string) []string {
	t.Helper()
	var result listResult
	require.NoError(t, json.Unmarshal([]byte(out), &result), "output should be valid JSON")

	names := make([]string, 0, len(result.Carriers))
	for _, carrier := range result.Carriers {
		names = append(names, carrier.Name)
	}
	return names
}

func addonNames(t *testing.T, out string) []string {
	t.Helper()
	var result addonsListResult
	require.NoError(t, json.Unmarshal([]byte(out), &result), "output should be valid JSON")
	return result.Addons
}

// On VentunoQ with Ubuntu the addons are available, together with the media carrier.
func TestVentunoqUbuntuBoardSupport(t *testing.T) {
	startVentunoqUbuntuDockerContainer(t)
	t.Cleanup(func() { stopVentunoqDockerContainer(t) })

	addons := addonNames(t, execInVentunoqContainer(t, "arduino-linux-config", "addons", "list", "--format", "json"))
	require.ElementsMatch(t, []string{"audio-codec-zero", "automation"}, addons)

	carriers := carrierNames(t, execInVentunoqContainer(t, "arduino-linux-config", "carrier", "list", "--format", "json"))
	require.Equal(t, []string{"media-carrier"}, carriers)
}

// On UnoQ the addons are not supported, so the command is not even registered.
func TestUnoqBoardSupport(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out, err := execInNamedContainerWithError(t, containerName, "arduino-linux-config", "addons", "list")
	require.Error(t, err, "the addons command should not be available on UnoQ")
	require.Contains(t, out, `unknown command "addons"`)

	carriers := carrierNames(t, execInContainer(t, "arduino-linux-config", "carrier", "list", "--format", "json"))
	require.Equal(t, []string{"media-carrier"}, carriers)
}

// On VentunoQ the only supported distribution is Ubuntu: on Debian neither
// carriers nor addons are available.
func TestVentunoqDebianBoardSupport(t *testing.T) {
	startVentunoqDebianDockerContainer(t)
	t.Cleanup(func() { stopVentunoqDockerContainer(t) })

	out, err := execInNamedContainerWithError(t, ventunoqContainerName, "arduino-linux-config", "carrier", "list", "--format", "json")
	require.Error(t, err)
	require.Contains(t, out, `unsupported board/os`)
}
