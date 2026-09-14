// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package hw

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// The carrier group is the old name of hw, so it offers the same commands and
// stays hidden on every one of them.
func TestHiddenCarrierMirrorsHwCmd(t *testing.T) {
	names := func(cmd *cobra.Command) []string {
		result := make([]string, 0, len(cmd.Commands()))
		for _, sub := range cmd.Commands() {
			result = append(result, sub.Name())
			require.True(t, sub.Hidden, "%s %s must be hidden", cmd.Name(), sub.Name())
			require.Empty(t, sub.Deprecated, "%s %s must not warn", cmd.Name(), sub.Name())
		}
		return result
	}

	carrierCmd := NewCarrierCmd()
	require.Equal(t, "carrier", carrierCmd.Name())
	require.True(t, carrierCmd.Hidden)
	require.Empty(t, carrierCmd.Deprecated)
	require.Equal(t, []string{"disable", "enable", "list", "reload", "show"}, names(carrierCmd))

	hwCmd := NewHwCmd()
	require.False(t, hwCmd.Hidden)
	for _, sub := range hwCmd.Commands() {
		require.False(t, sub.Hidden, "hw %s must not be hidden", sub.Name())
		require.Empty(t, sub.Deprecated, "hw %s must not warn", sub.Name())
	}
}
