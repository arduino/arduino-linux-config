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

func SetupDebian() func() {
	return setupOs("debian")
}

func SetupUbuntu() func() {
	return setupOs("ubuntu")
}

func SetupCompatUnoq() func() {
	return setupCompat("arduino,imola\x00")
}
func SetupCompatVentunoq() func() {
	return setupCompat("arduino,monza\x00")
}

func setupOs(osId string) func() {
	etcDir, err := os.MkdirTemp("", "etc-root")
	if err != nil {
		panic(err)
	}
	etcPath := filepath.Join(etcDir, "etc")
	if err := os.MkdirAll(etcPath, 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(etcPath, "os-release"), []byte(fmt.Sprintf("ID=%s\n", osId)), 0600); err != nil {
		panic(err)
	}
	os.Setenv("OS_RELEASE_FS_DIR", etcDir)
	return func() {
		os.Unsetenv("OS_RELEASE_FS_DIR")
		os.RemoveAll(etcDir)
	}
}

func setupCompat(board string) func() {
	compatDir, err := os.MkdirTemp("", "compat-root")
	if err != nil {
		panic(err)
	}
	compatPath := filepath.Join(compatDir, "sys", "firmware", "devicetree", "base")
	if err := os.MkdirAll(compatPath, 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(compatPath, "compatible"), []byte(board), 0600); err != nil {
		panic(err)
	}
	os.Setenv("COMPATIBLE_FS_DIR", compatDir)
	return func() {
		os.Unsetenv("COMPATIBLE_FS_DIR")
		os.RemoveAll(compatDir)
	}
}
