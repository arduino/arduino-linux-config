// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"
)

type Compatible []string

func GetBoardID() string {
	compatible := loadCompatible()
	slog.Debug("detected platform", "compatible", compatible)
	switch {
	case compatible.IsCompatibleWith("arduino,imola"):
		return "unoq"
	case compatible.IsCompatibleWith("arduino,monza"):
		return "ventunoq"
	default:
		slog.Warn("not supported platform", "compatible", compatible)
	}
	return ""
}

// reads the os from the root FS,
// or from COMPATIBLE_ROOT_DIR if set (used in integration tests).
func GetLinuxDistribution() string {
	root := "/"
	if dir := os.Getenv("COMPATIBLE_ROOT_DIR"); dir != "" {
		root = dir
	}
	return getLinuxDistributionFromFS(os.DirFS(root))
}

func getCompatibleFromFS(fs fs.FS) Compatible {
	var compatibles []string
	if comp, err := fs.Open("sys/firmware/devicetree/base/compatible"); err == nil {
		defer comp.Close()
		if data, err := io.ReadAll(comp); err == nil {
			for _, c := range bytes.Split(data, []byte{'\x00'}) {
				c = bytes.Trim(c, "\x00 \t\n\r") // trim null bytes and whitespace
				if len(c) > 0 {
					compatibles = append(compatibles, string(c))
				}
			}
		}
	}
	return compatibles
}

func getLinuxDistributionFromFS(fs fs.FS) string {
	if f, err := fs.Open("etc/os-release"); err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "ID=") {
				osId := strings.TrimPrefix(line, "ID=")
				return strings.Trim(osId, "\n\t\" ")
			}
		}
	}
	return ""
}

// loadCompatible reads the device-tree compatible strings from the root FS,
// or from COMPATIBLE_ROOT_DIR if set (used in integration tests).
func loadCompatible() Compatible {
	root := "/"
	if dir := os.Getenv("COMPATIBLE_ROOT_DIR"); dir != "" {
		root = dir
	}
	return getCompatibleFromFS(os.DirFS(root))
}

func (c Compatible) IsCompatibleWith(prefix string) bool {
	for _, comp := range c {
		if strings.HasPrefix(comp, prefix) {
			return true
		}
	}
	return false
}
