// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package reload

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/config"
	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/testutil"
)

func TestNewReloadCmd(t *testing.T) {
	t.Cleanup(testutil.SetupUnoQDebian())

	cmd := NewReloadCmd()
	require.Equal(t, "reload", cmd.Use)
	require.NotNil(t, cmd.Flags().Lookup("dry-run"))
}

func TestReloadHandlerBoards(t *testing.T) {
	tests := []struct {
		name  string
		setup func() func()
	}{
		{name: "unoq", setup: testutil.SetupUnoQDebian},
		{name: "ventunoq", setup: testutil.SetupVentunoQUbuntu},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(tt.setup())

			reg := registry.New()
			// dry-run avoids the root check and any state mutation.
			require.NotPanics(t, func() {
				reloadHandler(context.Background(), reg, config.New(), true)
			})
		})
	}
}

func TestReloadResultString(t *testing.T) {
	r := reloadResult{
		BoardID:  "unoq",
		DryRun:   true,
		Reloaded: []string{"media-carrier"},
	}
	out := r.String()
	require.True(t, strings.Contains(out, "unoq"))
	require.True(t, strings.Contains(out, "media-carrier"))
}
