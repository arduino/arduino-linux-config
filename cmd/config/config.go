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
	statusFile  *paths.Path
	factoryDTB  *paths.Path
	actualDTB   *paths.Path
	overlaysDir *paths.Path
}

func NewConfigFromEnv() (Configuration, error) {
	statusFile := paths.New("/var/lib/arduino-linux-config/status/status.json")
	baseDTB := paths.New("/boot/efi/dtb/qcom/qrb2210-arduino-imola-factory.dtb")
	actualDTB := paths.New("/boot/efi/dtb/qcom/qrb2210-arduino-imola.dtb")
	overlaysDir := paths.New("/boot/efi/dtb/qcom/")

	c := Configuration{
		statusFile:  statusFile,
		factoryDTB:  baseDTB,
		actualDTB:   actualDTB,
		overlaysDir: overlaysDir,
	}

	if err := c.init(); err != nil {
		return Configuration{}, err
	}

	return c, nil
}

func (c *Configuration) init() error {
	if c.factoryDTB.NotExist() {
		return fmt.Errorf("base DTB not found: %s", c.factoryDTB)
	}
	if c.overlaysDir.NotExist() {
		return fmt.Errorf("overlays directory not found: %s", c.overlaysDir)
	}
	return nil
}

func (c *Configuration) StatusFile() *paths.Path {
	return c.statusFile
}

func (c *Configuration) FactoryDTB() *paths.Path {
	return c.factoryDTB
}

func (c *Configuration) ActualDTB() *paths.Path {
	return c.actualDTB
}

func (c *Configuration) OverlaysDir() *paths.Path {
	return c.overlaysDir
}
