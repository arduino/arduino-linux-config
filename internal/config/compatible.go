// This file is part of arduino-app-cli.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	"github.com/arduino/go-paths-helper"
)

type Compatible []string

// loadCompatible reads the device-tree compatible strings from the root FS,
// or from COMPATIBLE_FS_DIR if set (used in integration tests).
func loadCompatible() Compatible {
	root := "/"
	if dir := os.Getenv("COMPATIBLE_FS_DIR"); dir != "" {
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

func GetLinuxDistribution() (string, error) {
	f, err := paths.New("/etc/os-release").ReadFile()
	if err != nil {
		return "", fmt.Errorf("failed to read os-release file: %w", err)
	}

	s := bufio.NewScanner(bytes.NewReader(f))
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "ID=") {
			prettyName := strings.TrimPrefix(line, "ID=")
			return strings.Trim(prettyName, "\n\t\" "), nil
		}
	}

	return "", fmt.Errorf("ID not found in os-release file")
}

func GetBoardID() (string, error) {
	compatible := loadCompatible()
	slog.Debug("detected platform", "compatible", compatible)
	switch {
	case compatible.IsCompatibleWith("arduino,imola"):
		return "unoq", nil
	case compatible.IsCompatibleWith("arduino,monza"):
		return "ventunoq", nil
	default:
		slog.Warn("not supported platform", "compatible", compatible)
	}
	return "", fmt.Errorf("failed to identify board id")
}
