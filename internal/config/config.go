// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"fmt"

	"github.com/arduino/arduino-linux-config/internal/dto"
	"github.com/arduino/go-paths-helper"
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

func GetBoard() (dto.Board, error) {
	switch GetBoardID() {
	case "unoq":
		return dto.UnoQ{}, nil
	case "ventunoq":
		return dto.VentunoQ{}, nil
	default:
		return nil, fmt.Errorf("unsupported board/os")
	}
}
