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
		switch GetLinuxDistribution() {
		case "ubuntu":
			baseDtbFullPath, err := deviceTreeDiscover()
			if err != nil {
				return nil, fmt.Errorf("failed to discover device tree: %w", err)
			}
			return dto.VentunoQ{
				BaseDtbFullPath: baseDtbFullPath,
				OverlaysDir:     paths.New("/var/lib/arduino-linux-config/overlays/"),
				DtbFileName:     filepath.Base(baseDtbFullPath),
			}, nil
		case "debian":
			return dto.VentunoQ{
				BaseDtbFileName: "combined-dtb-base.dtb",
				OverlaysDir:     paths.New("/var/lib/arduino-linux-config/overlays/"),
				DtbFileName:     "combined-dtb.dtb",
			}, nil
		}
	}
	return nil, fmt.Errorf("unsupported board/os")
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

	return filepath.Join(root, baseDtbFullPath), nil
}

var kernelVersionPattern = regexp.MustCompile(`^[1-9]\.[0-9]\.[0-9]-`)

// Prefers grub.cfg (the next kernel to boot); falls back to the one kernel
// under /boot on boards with no grub.cfg (e.g. a U-Boot/EFI boot flow).
func kernelVersionDiscover(root string) (string, error) {
	if version, err := kernelVersionFromGrub(root); err == nil {
		return version, nil
	}
	return kernelVersionFromBootDir(root)
}

func kernelVersionFromGrub(root string) (string, error) {
	grubConfig := filepath.Join(root, "boot/grub/grub.cfg")
	// #nosec G702 -- no shell or user input involved
	//nolint:forbidigo // grep only reads, no side effect for --dry-run to report.
	output, err := exec.Command("grep", "-m", "1", "/boot/vmlinuz-", grubConfig).Output()
	if err != nil {
		return "", err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return "", fmt.Errorf("error getting kernel version")
	}

	return strings.TrimPrefix(fields[1], "/boot/vmlinuz-"), nil
}

func kernelVersionFromBootDir(root string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(root, "boot/vmlinuz-*"))
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one /boot/vmlinuz-*, found %d", len(matches))
	}
	return strings.TrimPrefix(filepath.Base(matches[0]), "vmlinuz-"), nil
}

// implement the same behavior of /etc/kernel/postinst.d/zzz-update-dtb
func deviceTreeDiscoverFromFS(root fs.FS, version string) (string, error) {
	switch getLinuxDistributionFromFS(root) {
	case "ubuntu", "debian":
	default:
		return "", fmt.Errorf("unsupported distribution")
	}

	if !kernelVersionPattern.MatchString(version) {
		return "", fmt.Errorf("error getting kernel version")
	}

	for _, dtbPath := range []string{
		"lib/firmware/" + version + "/device-tree/qcom/combined-dtb.dtb",
		"usr/lib/linux-image-" + version + "/qcom/combined-dtb.dtb",
	} {
		if _, err := fs.Stat(root, dtbPath); err == nil {
			return "/" + dtbPath, nil
		}
	}

	return "", fmt.Errorf("no valid device tree found")
}
