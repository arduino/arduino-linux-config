// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package testutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func SetupUnoQDebian() func() {
	return envRootSetup("arduino,imola\x00", "debian")
}

func SetupVentunoQDebian() func() {
	return envRootSetup("arduino,monza\x00", "debian")
}

func SetupVentunoQUbuntu() func() {
	return envRootSetup("arduino,monza\x00", "ubuntu")
}

func envRootSetup(board, osId string) func() {
	compatRootDir, err := os.MkdirTemp("", "compat-root")
	if err != nil {
		panic(err)
	}
	compatPath := filepath.Join(compatRootDir, "sys", "firmware", "devicetree", "base")
	if err := os.MkdirAll(compatPath, 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(compatPath, "compatible"), []byte(board), 0600); err != nil {
		panic(err)
	}

	etcPath := filepath.Join(compatRootDir, "etc")
	if err := os.MkdirAll(etcPath, 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(etcPath, "os-release"), []byte(fmt.Sprintf("ID=%s\n", osId)), 0600); err != nil {
		panic(err)
	}

	bootIdPath := filepath.Join(compatRootDir, "proc", "sys", "kernel", "random")
	if err := os.MkdirAll(bootIdPath, 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(bootIdPath, "boot_id"), []byte("test-boot-id\n"), 0600); err != nil {
		panic(err)
	}

	os.Setenv("COMPATIBLE_ROOT_DIR", compatRootDir)
	return func() {
		os.Unsetenv("COMPATIBLE_ROOT_DIR")
		os.RemoveAll(compatRootDir)
	}
}
