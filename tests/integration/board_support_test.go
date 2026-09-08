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

// list prints every mount of the board, so the names are selected by kind.
func mountNames(t *testing.T, out string, kind string) []string {
	t.Helper()
	var result listResult
	require.NoError(t, json.Unmarshal([]byte(out), &result), "output should be valid JSON")

	names := make([]string, 0, len(result.Mounts))
	for _, mount := range result.Mounts {
		if mount.Kind == kind {
			names = append(names, mount.Name)
		}
	}
	return names
}

// On VentunoQ with Ubuntu the hats are available, together with the media carrier.
func TestVentunoqUbuntuBoardSupport(t *testing.T) {
	startVentunoqUbuntuDockerContainer(t)
	t.Cleanup(func() { stopVentunoqDockerContainer(t) })

	out := execInVentunoqContainer(t, "arduino-linux-config", "hw", "list", "--format", "json")

	require.ElementsMatch(t, []string{"audio-codec-zero", "automation"}, mountNames(t, out, "hat"))
	require.Equal(t, []string{"media-carrier"}, mountNames(t, out, "carrier"))
}

// On UnoQ there is no hat connector, so the registry declares no hat.
func TestUnoqBoardSupport(t *testing.T) {
	startDockerContainer(t)
	t.Cleanup(func() { stopDockerContainer(t) })

	out := execInContainer(t, "arduino-linux-config", "hw", "list", "--format", "json")

	require.Equal(t, []string{"media-carrier"}, mountNames(t, out, "carrier"))
	require.Empty(t, mountNames(t, out, "hat"), "UnoQ has no hat connector")
}

// On VentunoQ the only supported distribution is Ubuntu: on Debian no mount
// is available.
func TestVentunoqDebianBoardSupport(t *testing.T) {
	startVentunoqDebianDockerContainer(t)
	t.Cleanup(func() { stopVentunoqDockerContainer(t) })

	out, err := execInNamedContainerWithError(t, ventunoqContainerName, "arduino-linux-config", "hw", "list", "--format", "json")
	require.Error(t, err)
	require.Contains(t, out, `unsupported board/os`)
}
