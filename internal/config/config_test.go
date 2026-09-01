// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestDeviceTreeDiscoverFromFS(t *testing.T) {
	const version = "6.8.0-test"

	tests := []struct {
		name        string
		root        fstest.MapFS
		expected    string
		expectedErr string
	}{
		{
			name: "firmware directory",
			root: fstest.MapFS{
				"etc/os-release":     {Data: []byte("ID=ubuntu\n")},
				"boot/grub/grub.cfg": {Data: []byte("linux /boot/vmlinuz-" + version + " root=UUID=test\n")},
				"lib/firmware/6.8.0-test/device-tree/qcom/combined-dtb.dtb": {Data: []byte("dtb")},
			},
			expected: "/lib/firmware/6.8.0-test/device-tree/qcom/combined-dtb.dtb",
		},
		{
			name: "linux image directory",
			root: fstest.MapFS{
				"etc/os-release":     {Data: []byte("ID=ubuntu\n")},
				"boot/grub/grub.cfg": {Data: []byte("linux /boot/vmlinuz-" + version + " root=UUID=test\n")},
				"usr/lib/linux-image-6.8.0-test/qcom/combined-dtb.dtb": {Data: []byte("dtb")},
			},
			expected: "/usr/lib/linux-image-6.8.0-test/qcom/combined-dtb.dtb",
		},
		{
			name: "firmware directory takes precedence",
			root: fstest.MapFS{
				"etc/os-release":     {Data: []byte("ID=ubuntu\n")},
				"boot/grub/grub.cfg": {Data: []byte("linux /boot/vmlinuz-" + version + " root=UUID=test\n")},
				"lib/firmware/6.8.0-test/device-tree/qcom/combined-dtb.dtb": {Data: []byte("dtb")},
				"usr/lib/linux-image-6.8.0-test/qcom/combined-dtb.dtb":      {Data: []byte("dtb")},
			},
			expected: "/lib/firmware/6.8.0-test/device-tree/qcom/combined-dtb.dtb",
		},
		{
			name: "unsupported distribution",
			root: fstest.MapFS{
				"etc/os-release":     {Data: []byte("ID=debian\n")},
				"boot/grub/grub.cfg": {Data: []byte("linux /boot/vmlinuz-" + version + " root=UUID=test\n")},
				"lib/firmware/6.8.0-test/device-tree/qcom/combined-dtb.dtb": {Data: []byte("dtb")},
			},
			expectedErr: "Unsupported distribution",
		},
		{
			name: "device tree not found",
			root: fstest.MapFS{
				"etc/os-release":     {Data: []byte("ID=ubuntu\n")},
				"boot/grub/grub.cfg": {Data: []byte("linux /boot/vmlinuz-" + version + " root=UUID=test\n")},
			},
			expectedErr: "No valid device tree found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dtbFullPath, err := deviceTreeDiscoverFromFS(tt.root, version)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				require.Empty(t, dtbFullPath)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, dtbFullPath)
		})
	}
}
