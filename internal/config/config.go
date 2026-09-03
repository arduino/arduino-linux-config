// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"fmt"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/dto"
)

type Configuration struct {
	statusDir *paths.Path
}

func New() Configuration {
	return Configuration{
		statusDir: paths.New(RootDir(), "var/lib/arduino-linux-config/status"),
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
			BaseDtbPath: paths.New("/var/lib/arduino-linux-config/overlays/ventunoq/"),
			BaseDtbFile: "qualcomm_technologies_inc._monaco_monza_addons.bin",
			OverlaysDir: paths.New("/var/lib/arduino-linux-config/overlays/"),
			DtbFileName: "combined-dtb.dtb",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported board/os")
	}
}
