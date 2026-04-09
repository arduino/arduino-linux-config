// This file is part of arduino-linux-config.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-linux-config.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package config

import (
	"fmt"

	"github.com/arduino/go-paths-helper"
)

type Configuration struct {
	statusDir      *paths.Path
	carriersConfig *paths.Path
	systemDTB      *paths.Path
	baseDTB        *paths.Path
}

func NewConfigFromEnv() (Configuration, error) {
	statusFile := paths.New("/var/lib/arduino-linux-config/status")
	systemDTB := paths.New("/boot/efi/dtb/qcom/qrb2210-arduino-imola.dtb")
	baseDTB := paths.New("/boot/efi/dtb/qcom/qrb2210-arduino-imola-base.dtb")

	c := Configuration{
		statusDir: statusFile,
		systemDTB: systemDTB,
		baseDTB:   baseDTB,
	}

	if err := c.init(); err != nil {
		return Configuration{}, err
	}

	return c, nil
}

func (c *Configuration) init() error {
	if c.carriersConfig.NotExist() {
		return fmt.Errorf("carriers configuration directory not found: %s", c.carriersConfig)
	}
	return nil
}

func (c *Configuration) StatusDir() *paths.Path {
	return c.statusDir
}

func (c *Configuration) CarriersConfig() *paths.Path {
	return c.carriersConfig
}

func (c *Configuration) SystemDTB() *paths.Path {
	return c.systemDTB
}

func (c *Configuration) BaseDTB() *paths.Path {
	return c.baseDTB
}
