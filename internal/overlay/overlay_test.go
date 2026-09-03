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
	"github.com/arduino/arduino-linux-config/internal/testutil"
)

func mediaCarrier(t *testing.T) registry.Carrier {
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
