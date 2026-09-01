// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arduino/go-paths-helper"
	"go.bug.st/f"

	"github.com/arduino/arduino-linux-config/internal/dto"
)

type Configuration struct {
	statusDir *paths.Path
}

func New() Configuration {
	return Configuration{
		statusDir: paths.New("/var/lib/arduino-linux-config/status"),
	}
}

func (c *Configuration) StatusDir() *paths.Path {
	return c.statusDir
}

func GetBoard() (dto.DeviceTreeApplier, error) {
	switch GetBoardID() {
	case "unoq":
		return dto.UnoQ{
			BaseDtbFile: "qrb2210-arduino-imola-base.dtb",
			OverlaysDir: paths.New("/boot/efi/dtb/qcom/"),
			DtbFileName: "qrb2210-arduino-imola.dtb",
		}, nil
	case "ventunoq":
		return dto.VentunoQ{
			BaseDtbFullPath: f.Must(deviceTreeDiscover()),
			OverlaysDir:     paths.New("/var/lib/arduino-linux-config/overlays/"),
			DtbFileName:     "combined-dtb.dtb",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported board/os")
	}
}

func deviceTreeDiscover() (string, error) {
	root := "/"
	if dir := os.Getenv("COMPATIBLE_ROOT_DIR"); dir != "" {
		root = dir
	}

	version, err := kernelVersionDiscover(root)
	if err != nil {
		return "", err
	}

	baseDtbFullPath, err := deviceTreeDiscoverFromFS(os.DirFS(root), version)
	if err != nil {
		return "", err
	}

	return baseDtbFullPath, nil
}

var kernelVersionPattern = regexp.MustCompile(`^[1-9]\.[0-9]\.[0-9]-`)

// implement the same behaviour of /etc/kernel/postinst.d/zzz-update-dtb
func kernelVersionDiscover(root string) (string, error) {
	grubConfig := filepath.Join(root, "boot/grub/grub.cfg")
	output, err := exec.Command("grep", "-m", "1", "/boot/vmlinuz-", grubConfig).Output()
	if err != nil {
		return "", err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return "", fmt.Errorf("Error getting kernel version")
	}

	return strings.TrimPrefix(fields[1], "/boot/vmlinuz-"), nil
}

// implement the same behaviour of /etc/kernel/postinst.d/zzz-update-dtb
func deviceTreeDiscoverFromFS(root fs.FS, version string) (string, error) {
	if getLinuxDistributionFromFS(root) != "ubuntu" {
		return "", fmt.Errorf("Unsupported distribution")
	}

	if !kernelVersionPattern.MatchString(version) {
		return "", fmt.Errorf("Error getting kernel version")
	}

	for _, dtbPath := range []string{
		"lib/firmware/" + version + "/device-tree/qcom/combined-dtb.dtb",
		"usr/lib/linux-image-" + version + "/qcom/combined-dtb.dtb",
	} {
		if _, err := fs.Stat(root, dtbPath); err == nil {
			return "/" + dtbPath, nil
		}
	}

	return "", fmt.Errorf("No valid device tree found")
}
