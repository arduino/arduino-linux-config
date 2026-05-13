// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
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
