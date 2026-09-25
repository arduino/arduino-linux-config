// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package registry

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/testutil"
)

// The name alone selects a part, so a carrier and a hat must not share one.
func TestMountNamesAreUnique(t *testing.T) {
	for name, setup := range map[string]func() func(){
		"unoq":     testutil.SetupUnoQDebian,
		"ventunoq": testutil.SetupVentunoQUbuntu,
	} {
		t.Run(name, func(t *testing.T) {
			t.Cleanup(setup())

			seen := make(map[MountName]Kind)
			for _, mount := range New().Mounts {
				require.NotContains(t, seen, mount.Name, "duplicated mount name")
				require.NotEmpty(t, mount.Kind, "mount %s has no kind", mount.Name)
				seen[mount.Name] = mount.Kind
			}
		})
	}
}

func TestKernelVersionComparisons(t *testing.T) {
	// Check patch
	require.True(t, isVersionAtLeast("7.0.0-g122c2c22d838", "7.0.0-g122c2c22d838"))
	require.False(t, isVersionAtLeast("7.0.0-g122c2c22d838", "7.0.1-g122c2c22d838"))
	require.True(t, isVersionAtLeast("7.0.1-g122c2c22d838", "7.0.0-g122c2c22d838"))

	// Check last string
	require.True(t, isVersionAtLeast("7.0.0-g122c2c22d100", "7.0.0-g122c2c22d100"))
	require.True(t, isVersionAtLeast("7.0.0-g122c2c22d100", "7.0.0-g122c2c22d099"))
	require.False(t, isVersionAtLeast("7.0.0-g122c2c22d100", "7.0.0-g122c2c22d101"))

	// Check minor
	require.False(t, isVersionAtLeast("7.0.0-g122c2c22d838", "7.2.0-g122c2c22d838"))
	require.True(t, isVersionAtLeast("7.2.0-g122c2c22d838", "7.0.0-g122c2c22d838"))

	// Check major
	require.True(t, isVersionAtLeast("9.2.0-g122c2c22d838", "8.2.0-g122c2c22d838"))
	require.False(t, isVersionAtLeast("10.2.0-g122c2c22d838", "11.2.0-g122c2c22d838"))

	require.True(t, isVersionAtLeast("6.8.0-1078-qcom", "6.8.0-1078-qcom"))
	require.True(t, isVersionAtLeast("6.8.0-1099-qcom", "6.8.0-1087-qcom"))
	require.False(t, isVersionAtLeast("6.8.0-1087-qcom", "6.8.0-1099-qcom"))
	require.False(t, isVersionAtLeast("6.7.0-1078-qcom", "6.8.0-1078-qcom"))
}

func TestIsVersionAtLeast(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		minReq   string
		expected bool
	}{
		{"empty minReq is always satisfied", "6.8.0-1078-qcom", "", true},
		{"identical strings are always satisfied", "6.8.0-1078-qcom", "6.8.0-1078-qcom", true},
		{"empty current with non-empty minReq is unsatisfied", "", "6.8.0-1078-qcom", false},
		{"higher ABI satisfies lower requirement", "6.8.0-1084-qcom", "6.8.0-1078-qcom", true},
		{"lower ABI does not satisfy higher requirement", "6.8.0-1078-qcom", "6.8.0-1084-qcom", false},
		{"lower kernel minor does not satisfy higher requirement", "6.7.0-1078-qcom", "6.8.0-1078-qcom", false},
		// dpkg --compare-versions "6.8.0-999-qcom" gt "6.8.0-1000-qcom" agrees: false
		{"ABI digit-count boundary is compared numerically", "6.8.0-999-qcom", "6.8.0-1000-qcom", false},
		{"non-semver fallback uses lexical string comparison", "not-a-version", "also-not-a-version", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isVersionAtLeast(tt.current, tt.minReq))
		})
	}
}

func TestSupportedDropsUnsupportedMountsAndDevices(t *testing.T) {
	reg := Registry{Mounts: []Mount{
		{
			Name:      MediaCarrier,
			Kind:      KindCarrier,
			OsSupport: true,
			Devices: []Device{
				{Name: Display, DeviceType: DeviceTypeDisplay, OsSupport: true},
				{Name: Camera0, DeviceType: DeviceTypeCamera, OsSupport: false},
			},
		},
		{Name: Automation, Kind: KindHat, OsSupport: false},
	}}

	supported := reg.Supported()
	require.Len(t, supported.Mounts, 1)
	require.Equal(t, MediaCarrier, supported.Mounts[0].Name)
	require.Len(t, supported.Mounts[0].Devices, 1)
	require.Equal(t, Display, supported.Mounts[0].Devices[0].Name)
}

func TestSupportMatrixCoversRegistryDtboReferences(t *testing.T) {
	matrixByDtbo := make(map[string]DtboSupport, len(NewSupportMatrix().Support))
	for _, support := range NewSupportMatrix().Support {
		matrixByDtbo[support.Dtbo] = support
	}

	seen := make(map[string]struct{})
	for _, mount := range append([]Mount{unoqMediaCarrier}, append(ventunoqUbuntuHats, ventunoqMediaCarrier)...) {
		for _, dtbo := range mount.EnabledDtbos {
			seen[dtbo] = struct{}{}
		}
		for _, dtbo := range mount.DisabledDtbos {
			seen[dtbo] = struct{}{}
		}
		for _, device := range mount.Devices {
			for _, option := range device.Options {
				for _, dtbo := range option.DtboFiles {
					seen[dtbo] = struct{}{}
				}
				for _, dtbo := range option.IncompatibleDtbo {
					seen[dtbo] = struct{}{}
				}
			}
		}
	}

	var missing []string
	for dtbo := range seen {
		support, ok := matrixByDtbo[dtbo]
		if !ok || support.IsSupported == nil {
			missing = append(missing, dtbo)
		}
	}

	require.Empty(t, missing, "missing support matrix entries for DTBO references in registry: %v", missing)
}
