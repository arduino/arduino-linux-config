// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/dto"
	"github.com/arduino/arduino-linux-config/internal/testutil"
)

func TestGetBoardUsesAddonsOnlyOnVentunoQDebian(t *testing.T) {
	t.Run("ventunoq debian", func(t *testing.T) {
		t.Cleanup(testutil.SetupVentunoQDebian())
		board, err := GetBoard()
		require.NoError(t, err)
		require.IsType(t, dto.VentunoQAddon{}, board)
	})

	t.Run("ventunoq ubuntu", func(t *testing.T) {
		t.Cleanup(testutil.SetupVentunoQUbuntu())
		board, _ := GetBoard()
		_, isAddon := board.(dto.VentunoQAddon)
		require.False(t, isAddon)
	})

	t.Run("unoq debian", func(t *testing.T) {
		t.Cleanup(testutil.SetupUnoQDebian())
		board, err := GetBoard()
		require.NoError(t, err)
		require.IsType(t, dto.UnoQ{}, board)
	})
}
