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
