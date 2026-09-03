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

func mediaMount(t *testing.T) registry.Mount {
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

	got := sorted(CollectDisabled(mediaMount(t)))
	want := []string{"qrb2210-arduino-imola-video_sound-usbc.dtbo"}
	require.Equal(t, want, got)
}

func TestCollectForStatus(t *testing.T) {
	t.Cleanup(testutil.SetupUnoQDebian())
	carrier := mediaMount(t)

	tests := []struct {
		name    string
		current status.MountStatus
		want    []string
	}{
		{
			name:    "disabled falls back to base overlays",
			current: status.MountStatus{Enable: false},
			want:    []string{"qrb2210-arduino-imola-video_sound-usbc.dtbo"},
		},
		{
			name: "enabled collects the configured overlays",
			current: status.MountStatus{
				Enable: true,
				StatusDevices: []status.StatusDevice{
					{Device: "camera0", Option: "type1-2lanes"},
				},
			},
			want: []string{
				"qrb2210-arduino-imola-carrier-media-camera-imx219-csi0-2lanes.dtbo",
				"qrb2210-arduino-imola-carrier-media.dtbo",
				"qrb2210-arduino-imola-video_sound-usbc.dtbo",
			},
		},
		{
			name: "enabled with incompatible base overlay removed",
			current: status.MountStatus{
				Enable: true,
				StatusDevices: []status.StatusDevice{
					{Device: "display", Option: "8-dsi-touch-a"},
				},
			},
			want: []string{
				"qrb2210-arduino-imola-carrier-media-panel-8in_touch_a-dsi.dtbo",
				"qrb2210-arduino-imola-carrier-media.dtbo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortedFirst(CollectForStatus(carrier, tt.current))
			require.Equal(t, tt.want, got)
		})
	}
}

func sortedFirst(files []string, _ []string) []string {
	return sorted(files)
}
