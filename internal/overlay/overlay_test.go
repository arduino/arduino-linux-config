// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package overlay

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/registry"
	"github.com/arduino/arduino-linux-config/internal/status"
	"github.com/arduino/arduino-linux-config/internal/testutil"
)

func mediaCarrier(t *testing.T) registry.Mount {
	t.Helper()
	reg := registry.New()
	carrier, exists := reg.FindByName(string(registry.MediaCarrier))
	require.True(t, exists, "media-carrier not found in registry")
	return carrier
}

func sorted(files []string) []string {
	files = slices.Clone(files)
	slices.Sort(files)
	return slices.Compact(files)
}

func TestCollectDisabled(t *testing.T) {
	t.Cleanup(testutil.SetupUnoQDebian())

	got := sorted(CollectDisabled(mediaCarrier(t)))
	want := []string{"qrb2210-arduino-imola-video_sound-usbc.dtbo"}
	require.Equal(t, want, got)
}

// CollectForStatus replaces the per-kind helpers: it resolves any mount from
// its persisted status.
func TestCollectForStatusDisabled(t *testing.T) {
	t.Cleanup(testutil.SetupUnoQDebian())

	files, incompatible := CollectForStatus(mediaCarrier(t), status.MountStatus{Enable: false})

	require.Equal(t, CollectDisabled(mediaCarrier(t)), files)
	require.Empty(t, incompatible)
}

func TestCollectForStatusEnabled(t *testing.T) {
	t.Cleanup(testutil.SetupVentunoQUbuntu())

	files, incompatible := CollectForStatus(mediaCarrier(t), status.MountStatus{
		Enable:        true,
		StatusDevices: []status.StatusDevice{{Device: "display", Option: "8-dsi-touch-a"}},
	})

	require.Equal(t, []string{"monaco-monza-dsi-waveshare,8.0-dsi-touch-a.dtbo"}, files)
	require.Empty(t, incompatible)
}

// A hat has no device, so its overlays come from the registry alone.
func TestCollectForStatusHat(t *testing.T) {
	t.Cleanup(testutil.SetupVentunoQUbuntu())

	hat, exists := registry.New().FindByName(string(registry.Automation))
	require.True(t, exists, "automation not found in registry")

	files, incompatible := CollectForStatus(hat, status.MountStatus{Enable: true})

	require.Equal(t, []string{"monaco-monza-automation-hat.dtbo"}, files)
	require.Empty(t, incompatible)
}
