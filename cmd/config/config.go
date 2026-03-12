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
	"os"

	"github.com/arduino/go-paths-helper"
)

type Configuration struct {
	statusDir   *paths.Path
	baseDTB     *paths.Path
	actualDTB   *paths.Path
	overlaysDir *paths.Path
}

func NewConfigFromEnv() (Configuration, error) {
	statusDir := paths.New(getEnvOrDefault(
		"ALC_CARRIER_STATUS_DIR",
		"/boot/arduino/carrier/status",
	))

	baseDTB := paths.New(getEnvOrDefault(
		"ALC_CARRIER_BASE_DTB",
		"/boot/arduino/qrb2210-arduino-imola-base.dtb",
	))

	actualDTB := paths.New(getEnvOrDefault(
		"ALC_CARRIER_ACTUAL_DTB",
		"/boot/arduino/qrb2210-arduino-imola.dtb",
	))

	overlaysDir := paths.New(getEnvOrDefault(
		"ALC_CARRIER_OVERLAYS_DIR",
		"/boot/arduino/overlays",
	))

	c := Configuration{
		statusDir:   statusDir,
		baseDTB:     baseDTB,
		actualDTB:   actualDTB,
		overlaysDir: overlaysDir,
	}

	if err := c.init(); err != nil {
		return Configuration{}, err
	}

	return c, nil
}

func (c *Configuration) init() error {
	if c.baseDTB.NotExist() {
		return fmt.Errorf("base DTB not found: %s", c.baseDTB)
	}
	if c.overlaysDir.NotExist() {
		return fmt.Errorf("overlays directory not found: %s", c.overlaysDir)
	}
	if err := c.statusDir.MkdirAll(); err != nil {
		return fmt.Errorf("failed to create status directory %s: %w", c.statusDir, err)
	}
	return nil
}

func (c *Configuration) StatusDir() *paths.Path {
	return c.statusDir
}

func (c *Configuration) BaseDTB() *paths.Path {
	return c.baseDTB
}

func (c *Configuration) ActualDTB() *paths.Path {
	return c.actualDTB
}

func (c *Configuration) OverlaysDir() *paths.Path {
	return c.overlaysDir
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
