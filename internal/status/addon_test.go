// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package status

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"

	"github.com/arduino/arduino-linux-config/internal/executor"
	"github.com/arduino/arduino-linux-config/internal/registry"
)

func addonReg() registry.Registry {
	return registry.Registry{
		Addons: []registry.Addon{
			{Name: registry.AudioCodecZero},
			{Name: registry.Automation},
		},
	}
}

func TestResolveAddons(t *testing.T) {
	reg := addonReg()
	currentBootId := "123"
	prevBootId := "001"

	tests := []struct {
		name   string
		status *AddonsStatusFile
		want   map[registry.AddonName]AddonState
	}{
		{
			name: "next from a previous boot is promoted to current",
			status: &AddonsStatusFile{
				CurrentStatus: AddonsStatus{Addons: map[registry.AddonName]AddonStatusInfo{}},
				NextStatus: AddonsStatus{Addons: map[registry.AddonName]AddonStatusInfo{
					registry.Automation: {Enabled: true, CreatedAt: prevBootId},
				}},
			},
			want: map[registry.AddonName]AddonState{
				registry.AudioCodecZero: {Name: registry.AudioCodecZero, Current: false, Next: false},
				registry.Automation:     {Name: registry.Automation, Current: true, Next: true},
			},
		},
		{
			name: "next from the current boot stays pending",
			status: &AddonsStatusFile{
				CurrentStatus: AddonsStatus{Addons: map[registry.AddonName]AddonStatusInfo{}},
				NextStatus: AddonsStatus{Addons: map[registry.AddonName]AddonStatusInfo{
					registry.Automation: {Enabled: true, CreatedAt: currentBootId},
				}},
			},
			want: map[registry.AddonName]AddonState{
				registry.AudioCodecZero: {Name: registry.AudioCodecZero, Current: false, Next: false},
				registry.Automation:     {Name: registry.Automation, Current: false, Next: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			states := resolveAddons(tt.status, reg, currentBootId)
			require.Len(t, states, len(reg.Addons))
			for _, st := range states {
				require.Equal(t, tt.want[st.Name], st)
			}
		})
	}
}

func TestApplyAddonChangesProducesSingleFileWithAllAddons(t *testing.T) {
	reg := addonReg()
	bootId := "b1f2-boot-id"

	status := newAddonsStatusFile()
	// Enable a single addon; the file must still contain every registry addon.
	applyAddonChanges(status, reg, map[registry.AddonName]bool{registry.Automation: true}, bootId)

	require.Len(t, status.CurrentStatus.Addons, len(reg.Addons))
	require.Len(t, status.NextStatus.Addons, len(reg.Addons))

	// The freshly enabled addon is pending: current disabled, next enabled.
	require.False(t, status.CurrentStatus.Addons[registry.Automation].Enabled)
	require.True(t, status.NextStatus.Addons[registry.Automation].Enabled)
	// Untouched addons are present and disabled.
	require.False(t, status.NextStatus.Addons[registry.AudioCodecZero].Enabled)

	data, err := json.MarshalIndent(status, "", "    ")
	require.NoError(t, err)
	t.Logf("addons.json:\n%s", data)
}

func TestAddonsStatusFileRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	p := paths.New(filepath.Join(tmpDir, addonsStatusFileName+".json"))

	// A missing file loads as an empty (disabled) status with initialized maps.
	loaded, err := loadAddonsStatusFile(p)
	require.NoError(t, err)
	require.NotNil(t, loaded.CurrentStatus.Addons)
	require.NotNil(t, loaded.NextStatus.Addons)

	want := AddonsStatusFile{
		CurrentStatus: AddonsStatus{
			CreatedAt: "001",
			Addons: map[registry.AddonName]AddonStatusInfo{
				registry.Automation: {Enabled: false, CreatedAt: "001"},
			},
		},
		NextStatus: AddonsStatus{
			CreatedAt: "001",
			Addons: map[registry.AddonName]AddonStatusInfo{
				registry.Automation: {Enabled: true, CreatedAt: "001"},
			},
		},
	}
	require.NoError(t, saveAddonsStatusFile(executor.Real(), p, want))

	got, err := loadAddonsStatusFile(p)
	require.NoError(t, err)
	require.Equal(t, want, *got)
}
