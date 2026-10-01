// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKernelVersionDiscoverPrefersGrub(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "boot/grub"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "boot/grub/grub.cfg"),
		[]byte("linux /boot/vmlinuz-6.8.0-1087-qcom root=/dev/sda1\n"),
		0644,
	))
	// Present but must be ignored: grub.cfg wins when it exists.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "boot"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "boot/vmlinuz-9.9.9-other"), nil, 0644))

	version, err := kernelVersionDiscover(root)
	require.NoError(t, err)
	require.Equal(t, "6.8.0-1087-qcom", version)
}

// Boards without grub (e.g. VentunoQ's Debian image, which boots via U-Boot/EFI
// and ships no /boot/grub/grub.cfg) fall back to the single kernel under /boot.
func TestKernelVersionDiscoverFallsBackToBootDirWithoutGrub(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "boot"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "boot/vmlinuz-7.1.0-arduino1+"), nil, 0644))

	version, err := kernelVersionDiscover(root)
	require.NoError(t, err)
	require.Equal(t, "7.1.0-arduino1+", version)
}

func TestKernelVersionDiscoverFailsWithAmbiguousBootDir(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "boot"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "boot/vmlinuz-1.0.0-a"), nil, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "boot/vmlinuz-2.0.0-b"), nil, 0644))

	_, err := kernelVersionDiscover(root)
	require.Error(t, err)
}
